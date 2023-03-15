/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * logv2.go
 */

package hatchet

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	COLLSCAN   = "COLLSCAN"
	DOLLAR_CMD = "$cmd"
	LIMIT      = 100
	TOP_N      = 23
)

var instance *Logv2

// GetLogv2 returns Logv2 instance
func GetLogv2() *Logv2 {
	if instance == nil {
		instance = &Logv2{dbfile: SQLITE3_FILE}
	}
	return instance
}

// Logv2 keeps Logv2 object
type Logv2 struct {
	buildInfo   map[string]interface{}
	dbfile      string
	filename    string
	legacy      bool
	hatchetName string
	isDigest    bool
	s3client    *S3Client
	testing     bool //test mode
	totalLines  int
	user        string
	verbose     bool
	version     string
}

// Logv2Info stores logv2 struct
type Logv2Info struct {
	Attr      bson.D    `json:"attr" bson:"attr"`
	Component string    `json:"c" bson:"c"`
	Context   string    `json:"ctx" bson:"ctx"`
	ID        int       `json:"id" bson:"id"`
	Msg       string    `json:"msg" bson:"msg"`
	Severity  string    `json:"s" bson:"s"`
	Timestamp time.Time `json:"t" bson:"t"`

	Attributes Attributes
	Message    string // remaining legacy message
	Client     *RemoteClient
}

type Attributes struct {
	Command            map[string]interface{} `json:"command" bson:"command"`
	ErrMsg             string                 `json:"errMsg" bson:"errMsg"`
	Milli              int                    `json:"durationMillis" bson:"durationMillis"`
	NS                 string                 `json:"ns" bson:"ns"`
	OriginatingCommand map[string]interface{} `json:"originatingCommand" bson:"originatingCommand"`
	PlanSummary        string                 `json:"planSummary" bson:"planSummary"`
	Reslen             int                    `json:"reslen" bson:"reslen"`
	Type               string                 `json:"type" bson:"type"`
}

type RemoteClient struct {
	Accepted int    `json:"accepted"`
	Conns    int    `json:"conns"`
	Ended    int    `json:"ended"`
	IP       string `json:"value"`
	Port     string `json:"port"`

	Driver  string // driver name
	Version string // driver version
}

// OpStat stores performance data
type OpStat struct {
	AvgMilli     float64 `json:"avg_ms"`        // max millisecond
	Count        int     `json:"count"`         // number of ops
	Index        string  `json:"index"`         // index used
	MaxMilli     int     `json:"max_ms"`        // max millisecond
	Namespace    string  `json:"ns"`            // database.collectin
	Op           string  `json:"op"`            // count, delete, find, remove, and update
	QueryPattern string  `json:"query_pattern"` // query pattern
	Reslen       int     `json:"total_reslen"`  // total reslen
	TotalMilli   int     `json:"total_ms"`      // total milliseconds
}

type LegacyLog struct {
	Timestamp string `json:"date"`
	Severity  string `json:"severity"`
	Component string `json:"component"`
	Context   string `json:"context"`
	Message   string `json:"message"` // remaining legacy message
}

type HatchetInfo struct {
	Arch    string
	End     string
	Module  string
	Name    string
	OS      string
	Start   string
	Version string

	Drivers  []map[string]string
	Provider string
	Region   string
}

// Analyze analyzes logs from a file
func (ptr *Logv2) Analyze(filename string) error {
	var err error
	var buf []byte
	var file *os.File
	var reader *bufio.Reader
	ptr.filename = filename
	ptr.hatchetName = getHatchetName(ptr.filename)
	if !ptr.legacy {
		log.Println("processing", filename)
		log.Println("hatchet name is", ptr.hatchetName)
	}

	if ptr.s3client != nil {
		var buf []byte
		toks := strings.Split(filename, "/")
		bucketName := toks[0]
		keyName := strings.Join(toks[1:], "/")
		if buf, err = ptr.s3client.GetObject(bucketName, keyName); err != nil {
			return err
		}
		if reader, err = GetBufioReader(buf); err != nil {
			return err
		}
		if !ptr.legacy {
			log.Println("s3 bucket", bucketName, "key", keyName)
		}
	} else if strings.HasPrefix(filename, "http://") || strings.HasPrefix(filename, "https://") {
		var username, password string
		if ptr.user != "" {
			toks := strings.Split(ptr.user, ":")
			if len(toks) == 2 {
				username = toks[0]
				password = toks[1]
			}
		}
		if ptr.isDigest {
			if reader, err = GetHTTPDigestContent(filename, username, password); err != nil {
				return err
			}
		} else {
			if reader, err = GetHTTPContent(filename, username, password); err != nil {
				return err
			}
		}
	} else {
		dirname := filepath.Dir(ptr.dbfile)
		os.Mkdir(dirname, 0755)
		if file, err = os.Open(filename); err != nil {
			return err
		}
		defer file.Close()
		if reader, err = gox.NewReader(file); err != nil {
			return err
		}
	}

	for { // check if it is in the logv2 format
		if buf, _, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
			return errors.New("no valid log format found")
		}
		if len(buf) == 0 {
			continue
		}
		str := string(buf)
		if !regexp.MustCompile("^{.*}$").MatchString(str) {
			return errors.New("log format not supported")
		}
		break
	}

	if !ptr.legacy && ptr.s3client != nil {
		// get total counts of logs
		log.Println("fast counting", filename, "...")
		ptr.totalLines, _ = gox.CountLines(reader)
		file.Seek(0, 0)
		if reader, err = gox.NewReader(file); err != nil {
			return err
		}
	}

	var isPrefix bool
	var stat *OpStat
	index := 0
	var start, end string
	var dbase Database

	if !ptr.legacy {
		if dbase, err = GetDatabase(ptr.hatchetName); err != nil {
			return err
		}
		defer dbase.Close()
		if err = dbase.Begin(); err != nil {
			return err
		}
	}

	for {
		if !ptr.testing && !ptr.legacy && index%50 == 0 && ptr.totalLines > 0 {
			fmt.Fprintf(os.Stderr, "\r%3d%% \r", (100*index)/ptr.totalLines)
		}
		if buf, isPrefix, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
			break
		}
		index++
		if len(buf) == 0 {
			continue
		}
		str := string(buf)
		for isPrefix {
			var bbuf []byte
			if bbuf, isPrefix, err = reader.ReadLine(); err != nil {
				break
			}
			str += string(bbuf)
		}

		doc := Logv2Info{}
		if err = bson.UnmarshalExtJSON([]byte(str), false, &doc); err != nil {
			continue
		}

		if err = AddLegacyString(&doc); err != nil {
			continue
		}
		if ptr.buildInfo == nil && doc.Msg == "Build Info" {
			ptr.buildInfo = doc.Attr.Map()["buildInfo"].(bson.D).Map()
		}
		if ptr.legacy {
			dt := getDateTimeStr(doc.Timestamp)
			logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", dt,
				doc.Severity, doc.Component, doc.Context, doc.Message)
			if !ptr.testing {
				fmt.Println(logstr)
			}
			continue
		}
		stat, _ = AnalyzeSlowOp(&doc)
		end = getDateTimeStr(doc.Timestamp)
		if start == "" {
			start = end
		}
		dbase.InsertLog(index, end, &doc, stat)
		if doc.Client != nil {
			if (doc.Client.Accepted + doc.Client.Ended) > 0 { // record connections
				dbase.InsertClientConn(index, &doc)
			} else if doc.Client.Driver != "" {
				if isAppDriver(doc.Client) {
					dbase.InsertDriver(index, &doc)
				}
			}
		}
	}
	if ptr.legacy {
		return nil
	}
	if err = dbase.Commit(); err != nil {
		return err
	}
	info := HatchetInfo{Start: start, End: end}
	if ptr.buildInfo != nil {
		if ptr.buildInfo["environment"] != nil {
			env := ptr.buildInfo["environment"].(bson.D).Map()
			info.Arch, _ = env["distarch"].(string)
			info.OS, _ = env["distmod"].(string)
		}
		if modules, ok := ptr.buildInfo["modules"].(bson.A); ok {
			if len(modules) > 0 {
				info.Module, _ = modules[0].(string)
			}
		}
		info.Version, _ = ptr.buildInfo["version"].(string)
	}
	if err = dbase.UpdateHatchetInfo(info); err != nil {
		return err
	}
	if err = dbase.CreateMetaData(); err != nil {
		return err
	}
	if !ptr.testing && !ptr.legacy {
		fmt.Fprintf(os.Stderr, "\r                         \r")
	}
	return ptr.PrintSummary()
}

func (ptr *Logv2) PrintSummary() error {
	dbase, err := GetDatabase(ptr.hatchetName)
	if err != nil {
		return err
	}
	fmt.Println("\n", "*", GetHatchetSummary(dbase.GetHatchetInfo()))
	summaries := []string{}
	var buffer bytes.Buffer
	buffer.WriteString("\r+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	buffer.WriteString(fmt.Sprintf("| Command  |COLLSCAN|avg ms| max ms | Count| %-32s| %-60s |\n", "Namespace", "Query Pattern"))
	buffer.WriteString("|----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------|\n")
	var ops []OpStat
	if ops, err = dbase.GetSlowOps("avg_ms", "DESC", false); err != nil {
		return err
	}
	for count, value := range ops {
		if count > TOP_N {
			break
		}
		str := value.QueryPattern
		if len(value.Op) > 10 {
			value.Op = value.Op[:10]
		}
		if len(value.Namespace) > 33 {
			length := len(value.Namespace)
			value.Namespace = value.Namespace[:1] + "*" + value.Namespace[(length-31):]
		}
		if len(str) > 60 {
			str = value.QueryPattern[:60]
			idx := strings.LastIndex(str, " ")
			if idx > 0 {
				str = value.QueryPattern[:idx]
			}
		}
		output := ""
		collscan := ""
		if value.Index == COLLSCAN {
			collscan = COLLSCAN
		}
		output = fmt.Sprintf("|%-10s %8s %6d %8d %6d %-33s %-62s|\n", value.Op, collscan,
			int(value.AvgMilli), value.MaxMilli, value.Count, value.Namespace, str)
		buffer.WriteString(output)
		if len(value.QueryPattern) > 60 {
			remaining := value.QueryPattern[len(str):]
			for i := 0; i < len(remaining); i += 60 {
				epos := i + 60
				var pstr string
				if epos > len(remaining) {
					epos = len(remaining)
					pstr = remaining[i:epos]
				} else {
					str = strings.Trim(remaining[i:epos], " ")
					idx := strings.LastIndex(str, " ")
					if idx >= 0 {
						pstr = str[:idx]
						i -= (60 - idx)
					} else {
						pstr = str
						i -= (60 - len(str))
					}
				}
				output = fmt.Sprintf("|%74s   %-62s|\n", " ", pstr)
				buffer.WriteString(output)
			}
		}
		if value.Index != "" && value.Index != COLLSCAN {
			output = fmt.Sprintf("|...index:  %-128s|\n", value.Index)
			buffer.WriteString(output)
		}
	}
	buffer.WriteString("+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	summaries = append(summaries, buffer.String())
	if TOP_N < len(ops) {
		summaries = append(summaries,
			fmt.Sprintf(` * %v: slowest %d of %d ops displayed`, ptr.hatchetName, TOP_N, len(ops)))
	}
	fmt.Println(strings.Join(summaries, "\n"))
	return err
}

func isAppDriver(client *RemoteClient) bool {
	driver := client.Driver
	version := client.Version

	if driver == "NetworkInterfaceTL" || driver == "MongoDB Internal Client" {
		return false
	} else if driver == "mongo-go-driver" && strings.HasSuffix(version, "-cloud") {
		return false
	}
	return true
}
