/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * logv2.go
 */

package hatchet

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
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
		instance = &Logv2{url: SQLITE3_FILE}
	}
	return instance
}

// Logv2 keeps Logv2 object
type Logv2 struct {
	buildInfo   map[string]interface{}
	logname     string
	legacy      bool
	hatchetName string
	isDigest    bool
	s3client    *S3Client
	testing     bool //test mode
	totalLines  int
	url         string // connection string
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
	Accepted int    `json:"accepted" bson:"accepted"`
	Conns    int    `json:"conns" bson:"conns"`
	Ended    int    `json:"ended" bson:"ended"`
	IP       string `json:"value" bson:"ip"`
	Port     string `json:"port" bson:"port"`

	Driver  string `bsno:"driver"`  // driver name
	Version string `bsno:"version"` // driver version
}

// OpStat stores performance data
type OpStat struct {
	AvgMilli     float64 `json:"avg_ms" bson:"avg_ms"`               // avg millisecond
	Count        int     `json:"count" bson:"count"`                 // number of ops
	Index        string  `json:"index" bson:"index"`                 // index used
	MaxMilli     int     `json:"max_ms" bson:"max_ms"`               // max millisecond
	Namespace    string  `json:"ns" bson:"ns"`                       // database.collectin
	Op           string  `json:"op" bson:"op"`                       // count, delete, find, remove, and update
	QueryPattern string  `json:"query_pattern" bson:"query_pattern"` // query pattern
	Reslen       int     `json:"total_reslen" bson:"total_reslen"`   // total reslen
	TotalMilli   int     `json:"total_ms" bson:"total_ms"`           // total milliseconds
}

type LegacyLog struct {
	Timestamp string `json:"date" bson:"date"`
	Severity  string `json:"severity" bson:"severity"`
	Component string `json:"component" bson:"component"`
	Context   string `json:"context" bson:"context"`
	Message   string `json:"message" bson:"message"` // remaining legacy message
}

type HatchetInfo struct {
	Arch    string `bson:"arch"`
	End     string `bson:"end"`
	Module  string `bson:"module"`
	Name    string `bson:"name"`
	OS      string `bson:"os"`
	Start   string `bson:"start"`
	Version string `bson:"version"`

	Drivers  []map[string]string
	Provider string `bson:"region"`
	Region   string `bson:"provider"`
}

func (ptr *Logv2) GetDBType() int {
	if strings.HasPrefix(ptr.url, "mongodb://") || strings.HasPrefix(ptr.url, "mongodb+srv://") {
		return Mongo
	}
	return SQLite3
}

// Analyze analyzes logs from a file
func (ptr *Logv2) Analyze(logname string) error {
	var err error
	var buf []byte
	var file *os.File
	var reader *bufio.Reader
	ptr.logname = logname
	ptr.hatchetName = getHatchetName(ptr.logname)
	if !ptr.legacy {
		log.Println("processing", logname)
		log.Println("hatchet name is", ptr.hatchetName)
	}

	if ptr.s3client != nil {
		var buf []byte
		if buf, err = ptr.s3client.GetObject(logname); err != nil {
			return err
		}
		if reader, err = GetBufioReader(buf); err != nil {
			return err
		}
	} else if strings.HasPrefix(logname, "http://") || strings.HasPrefix(logname, "https://") {
		var username, password string
		if ptr.user != "" {
			toks := strings.Split(ptr.user, ":")
			if len(toks) == 2 {
				username = toks[0]
				password = toks[1]
			}
		}
		if ptr.isDigest {
			if reader, err = GetHTTPDigestContent(logname, username, password); err != nil {
				return err
			}
		} else {
			if reader, err = GetHTTPContent(logname, username, password); err != nil {
				return err
			}
		}
	} else {
		if file, err = os.Open(logname); err != nil {
			return err
		}
		defer file.Close()
		if reader, err = gox.NewReader(file); err != nil {
			return err
		}
		if !ptr.legacy {
			log.Println("fast counting", logname, "...")
			ptr.totalLines, _ = gox.CountLines(reader)
			log.Println("counted", ptr.totalLines, "lines")
			if _, err = file.Seek(0, 0); err != nil {
				return err
			}
			if reader, err = gox.NewReader(file); err != nil {
				return err
			}
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
			log.Println("line", index, err)
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
	log.Println(GetHatchetSummary(dbase.GetHatchetInfo()))
	summaries := []string{}
	var buffer bytes.Buffer
	buffer.WriteString("\r+----------+--------+------+------+---------------------------------+----------------------------------------------------+\n")
	buffer.WriteString(fmt.Sprintf("| Command  |COLLSCAN|avg ms| Count| %-32s| %-50s |\n", "Namespace", "Query Pattern"))
	buffer.WriteString("|----------+--------+------+------+---------------------------------+----------------------------------------------------|\n")
	var ops []OpStat
	if ops, err = dbase.GetSlowOps("avg_ms", "DESC", false); err != nil {
		return err
	}
	lines := 5
	for count, value := range ops {
		if count > lines {
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
			str = value.QueryPattern[:50]
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
		output = fmt.Sprintf("|%-10s %8s %6d %6d %-33s %-52s|\n", value.Op, collscan,
			int(value.AvgMilli), value.Count, value.Namespace, str)
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
				output = fmt.Sprintf("|%74s   %-52s|\n", " ", pstr)
				buffer.WriteString(output)
			}
		}
		if value.Index != "" && value.Index != COLLSCAN {
			output = fmt.Sprintf("|...index: %-110s|\n", value.Index)
			buffer.WriteString(output)
		}
	}
	buffer.WriteString("+----------+--------+------+------+---------------------------------+----------------------------------------------------+")
	summaries = append(summaries, buffer.String())
	if lines < len(ops) {
		summaries = append(summaries,
			fmt.Sprintf(`+ %v: slowest %d of %d ops displayed`, ptr.hatchetName, lines+1, len(ops)))
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
