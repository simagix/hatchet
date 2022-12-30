// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"bufio"
	"bytes"
	"database/sql"
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
)

var instance *Logv2

// GetLogv2 returns Logv2 instance
func GetLogv2() *Logv2 {
	return instance
}

// Logv2 keeps Logv2 object
type Logv2 struct {
	buildInfo  map[string]interface{}
	dbfile     string
	filename   string
	legacy     bool
	tableName  string
	totalLines int
	verbose    bool
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
}

type Attributes struct {
	Command            map[string]interface{} `json:"command" bson:"command"`
	Milli              int                    `json:"durationMillis" bson:"durationMillis"`
	NS                 string                 `json:"ns" bson:"ns"`
	OriginatingCommand map[string]interface{} `json:"originatingCommand" bson:"originatingCommand"`
	PlanSummary        string                 `json:"planSummary" bson:"planSummary"`
	Reslen             int                    `json:"reslen" bson:"reslen"`
	Type               string                 `json:"type" bson:"type"`
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

// Analyze analyzes logs from a file
func (ptr *Logv2) Analyze(filename string) error {
	var err error
	var buf []byte
	ptr.filename = filename

	var file *os.File
	var reader *bufio.Reader
	ptr.filename = filename
	log.Println("processing", filename)
	ptr.tableName = "hatchet"
	dirname := filepath.Dir(ptr.dbfile)
	os.Mkdir(dirname, 0755)
	temp := filepath.Base(ptr.filename)
	i := strings.LastIndex(temp, ".log")
	if i > 0 {
		hatchetName := getTableName(temp[0:i])
		log.Println("hatchet table is", hatchetName)
		ptr.tableName = hatchetName
	}
	if file, err = os.Open(filename); err != nil {
		return err
	}
	defer file.Close()
	if reader, err = gox.NewReader(file); err != nil {
		return err
	}

	// check if it's a ptr format
	if buf, _, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
		return err
	}

	if len(buf) == 0 {
		return errors.New("can't detect file type")
	}

	str := string(buf)
	if !regexp.MustCompile("^{.*}$").MatchString(str) {
		return errors.New("log format not supported")
	}

	// get total counts of logs
	if !ptr.legacy {
		ptr.totalLines, _ = gox.CountLines(reader)
	}
	file.Seek(0, 0)
	if reader, err = gox.NewReader(file); err != nil {
		return err
	}

	var isPrefix bool
	var stat OpStat
	index := 0
	var sqlStmt string
	var pstmt *sql.Stmt
	var tx *sql.Tx

	if !ptr.legacy {
		db, err := sql.Open("sqlite3", ptr.dbfile)
		if err != nil {
			return err
		}
		defer db.Close()

		log.Println("creating table", ptr.tableName)
		sqlStmt = getHatchetInitStmt(ptr.tableName)
		if _, err = db.Exec(sqlStmt); err != nil {
			return err
		}

		tx, err = db.Begin()
		if err != nil {
			return err
		}
		pstmt, err = tx.Prepare(getHatchetPreparedStmt(ptr.tableName))
		if err != nil {
			return err
		}
		defer pstmt.Close()
	}

	for {
		if !ptr.verbose && !ptr.legacy && index%50 == 0 {
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
			logstr := fmt.Sprintf("%v.000Z %-2s %-8s [%v] %v", doc.Timestamp.Format(time.RFC3339)[:19],
				doc.Severity, doc.Component, doc.Context, doc.Message)
			fmt.Println(logstr)
		}

		if doc.Attr.Map()["command"] != nil {
			doc.Attributes.Command = doc.Attr.Map()["command"].(bson.D).Map()
		}
		if doc.Attr.Map()["ns"] != nil {
			doc.Attributes.NS = doc.Attr.Map()["ns"].(string)
		}
		if doc.Attr.Map()["durationMillis"] != nil {
			doc.Attributes.Milli = ToInt(doc.Attr.Map()["durationMillis"])
		}
		if doc.Attr.Map()["planSummary"] != nil {
			doc.Attributes.PlanSummary = doc.Attr.Map()["planSummary"].(string)
		}
		if doc.Attr.Map()["reslen"] != nil {
			doc.Attributes.Reslen = ToInt(doc.Attr.Map()["reslen"])
		}
		if doc.Attr.Map()["type"] != nil {
			doc.Attributes.Type = doc.Attr.Map()["type"].(string)
		}

		if stat, err = AnalyzeSlowOp(&doc); err != nil {
			stat = OpStat{}
		}
		if !ptr.legacy {
			if _, err = pstmt.Exec(index, doc.Timestamp.Format(time.RFC3339), doc.Severity, doc.Component, doc.Context,
				doc.Msg, doc.Attributes.PlanSummary, doc.Attr.Map()["type"], doc.Attr.Map()["ns"], doc.Message,
				stat.Op, stat.QueryPattern, stat.Index, doc.Attributes.Milli, doc.Attributes.Reslen); err != nil {
				return err
			}
		}
	}
	if ptr.legacy {
		return nil
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if !ptr.verbose && !ptr.legacy {
		fmt.Fprintf(os.Stderr, "\r                         \r")
	}

	return ptr.PrintSummary()
}

func (ptr *Logv2) PrintSummary() error {
	if ptr.buildInfo != nil {
		fmt.Println(gox.Stringify(ptr.buildInfo, "", "  "))
	}
	str, err := ptr.GetSlowOpsStats()
	if err != nil {
		return err
	}
	fmt.Println(str)
	return err
}

// GetSlowOpsStats prints slow ops stats
func (ptr *Logv2) GetSlowOpsStats() (string, error) {
	var maxSize = 20
	summaries := []string{}
	var buffer bytes.Buffer
	buffer.WriteString("\r+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	buffer.WriteString(fmt.Sprintf("| Command  |COLLSCAN|avg ms| max ms | Count| %-32s| %-60s |\n", "Namespace", "Query Pattern"))
	buffer.WriteString("|----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------|\n")

	ops, err := getSlowOps(ptr.tableName, "avg_ms", "DESC", false)
	if err != nil {
		return "", err
	}
	for count, value := range ops {
		if count > maxSize {
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
		if value.Index == "COLLSCAN" {
			collscan = "COLLSCAN"
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
		if value.Index != "" && value.Index != "COLLSCAN" {
			output = fmt.Sprintf("|...index:  %-128s|\n", value.Index)
			buffer.WriteString(output)
		}
	}
	buffer.WriteString("+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	summaries = append(summaries, buffer.String())
	if maxSize < len(ops) {
		summaries = append(summaries, fmt.Sprintf(`top %d of %d lines displayed.`, maxSize, len(ops)))
	}
	summaries = append(summaries, "hatchet table is "+ptr.tableName)
	return strings.Join(summaries, "\n"), err
}
