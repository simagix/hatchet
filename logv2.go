// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/simagix/gox"
)

const (
	COLLSCAN    = "COLLSCAN"
	DOLLAR_CMD  = "$cmd"
	SQLITE_FILE = "./data/hatchet.db"
)

// Logv2 keeps Logv2 object
type Logv2 struct {
	buildInfo  map[string]interface{}
	filename   string
	legacy     bool
	tableName  string
	totalLines int
	verbose    bool
}

// Logv2Info stores logv2 struct
type Logv2Info struct {
	Attr      map[string]interface{} `json:"attr" bson:"attr"`
	Component string                 `json:"c" bson:"c"`
	Context   string                 `json:"ctx" bson:"ctx"`
	ID        int                    `json:"id" bson:"id"`
	Msg       string                 `json:"msg" bson:"msg"`
	Severity  string                 `json:"s" bson:"s"`
	Timestamp map[string]string      `json:"t" bson:"t"`

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

// OpPattern stores performance data
type OpPattern struct {
	Command   string `bson:"command"` // count, delete, find, remove, and update
	Filter    string `bson:"filter"`  // query pattern
	Index     string `bson:"index"`   // index used
	Milli     int    `bson:"milli"`   // max millisecond
	Namespace string `bson:"ns"`      // database.collectin
	Plan      string `bson:"plan"`
	Reslen    int64  `bson:"reslen"` // total reslen

	Count       int   `bson:"count"`       // number of ops
	MaxMilli    int   `bson:"maxmilli"`    // max millisecond
	TotalMilli  int64 `bson:"totalmilli"`  // total milliseconds
	TotalReslen int64 `bson:"totalreslen"` // total reslen
}

// LogStats log stats structure
type LogStats struct {
	filter string
	index  string
	milli  int
	ns     string
	op     string
	reslen int
	date   string
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
		log.Fatal("log format not supported")
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
	var stat LogStats
	index := 0

	db, err := sql.Open("sqlite3", SQLITE_FILE)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("creating table", ptr.tableName)
	db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %v;`, ptr.tableName))
	sqlStmt := fmt.Sprintf(`CREATE TABLE %v 
		(	id integer not null primary key, date text, severity text, component text, context text,
			msg text, plan text, type text, ns text, message text,
			op text, filter text, _index text, milli integer, reslen integer);
		CREATE INDEX idx_op ON %v (op,ns,filter)`, ptr.tableName, ptr.tableName)
	if _, err = db.Exec(sqlStmt); err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
	}
	pstmt, err := tx.Prepare(fmt.Sprintf(`INSERT INTO 
		%v(	id, date, severity, component, context,
            msg, plan, type, ns, message, op, filter, _index, milli, reslen)
        VALUES(?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?)`, ptr.tableName))
	if err != nil {
		log.Fatal(err)
	}
	defer pstmt.Close()

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
		if err = json.Unmarshal([]byte(str), &doc); err != nil {
			continue
		}

		if err = AddLegacyString(&doc); err != nil {
			continue
		}
		if ptr.buildInfo == nil && doc.Msg == "Build Info" {
			ptr.buildInfo = doc.Attr["buildInfo"].(map[string]interface{})
		}

		if ptr.legacy {
			logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", doc.Timestamp["$date"], doc.Severity, doc.Component, doc.Context, doc.Message)
			fmt.Println(logstr)
		}

		if doc.Attr["command"] != nil {
			doc.Attributes.Command = doc.Attr["command"].(map[string]interface{})
		}
		if doc.Attr["ns"] != nil {
			doc.Attributes.NS = doc.Attr["ns"].(string)
		}
		if doc.Attr["durationMillis"] != nil {
			doc.Attributes.Milli = ToInt(doc.Attr["durationMillis"])
		}
		if doc.Attr["planSummary"] != nil {
			doc.Attributes.PlanSummary = doc.Attr["planSummary"].(string)
		}
		if doc.Attr["reslen"] != nil {
			doc.Attributes.Reslen = ToInt(doc.Attr["reslen"])
		}
		if doc.Attr["type"] != nil {
			doc.Attributes.Type = doc.Attr["type"].(string)
		}

		if stat, err = AnalyzeSlowQuery(&doc); err != nil {
			stat = LogStats{}
		}
		if _, err = pstmt.Exec(index, doc.Timestamp["$date"], doc.Severity, doc.Component, doc.Context,
			doc.Msg, doc.Attr["planSummary"], doc.Attr["type"], doc.Attr["ns"], doc.Message,
			stat.op, stat.filter, stat.index, doc.Attr["durationMillis"], doc.Attr["reslen"]); err != nil {
			log.Fatal(err)
		}
	}
	if ptr.legacy {
		return nil
	}
	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}
	if !ptr.verbose && !ptr.legacy {
		fmt.Fprintf(os.Stderr, "\r                         \r")
	}

	ptr.PrintSummary()
	return nil
}

func (ptr *Logv2) PrintSummary() {
	if ptr.buildInfo != nil {
		fmt.Println(gox.Stringify(ptr.buildInfo, "", "  "))
	}
	fmt.Println(ptr.GetSlowQueyiesSummary())
}

// printLogsSummary prints loginfo summary
func (ptr *Logv2) GetSlowQueyiesSummary() string {
	var maxSize = 20
	red := ""
	green := ""
	tail := ""
	summaries := []string{}
	var buffer bytes.Buffer
	buffer.WriteString("\r+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	buffer.WriteString(fmt.Sprintf("| Command  |COLLSCAN|avg ms| max ms | Count| %-32s| %-60s |\n", "Namespace", "Query Pattern"))
	buffer.WriteString("|----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------|\n")
	count := 0

	db, err := sql.Open("sqlite3", SQLITE_FILE)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	query := fmt.Sprintf(`SELECT op, COUNT(*) "count", ROUND(AVG(milli),1) avg_ms, MAX(milli) max_ms, SUM(milli) total_ms,
			ns, _index "index", SUM(reslen) "reslen", filter "query pattern"
			FROM %v WHERE op != "" GROUP BY op, ns, filter ORDER BY avg_ms DESC`, ptr.tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		count++
		if count >= maxSize {
			continue
		}
		value := OpPattern{}
		var avg float64
		if err = rows.Scan(&value.Command, &value.Count, &avg, &value.MaxMilli, &value.TotalMilli,
			&value.Namespace, &value.Index, &value.TotalReslen, &value.Filter); err != nil {
			log.Fatal(err)
		}
		str := value.Filter
		if len(value.Command) > 10 {
			value.Command = value.Command[:10]
		}
		if len(value.Namespace) > 33 {
			length := len(value.Namespace)
			value.Namespace = value.Namespace[:1] + "*" + value.Namespace[(length-31):]
		}
		if len(str) > 60 {
			str = value.Filter[:60]
			idx := strings.LastIndex(str, " ")
			if idx > 0 {
				str = value.Filter[:idx]
			}
		}
		output := ""
		if value.Plan == COLLSCAN {
			output = fmt.Sprintf("|%-10s %v%8s%v %6d %8d %6d %-33s %v%-62s%v|\n", value.Command, red, value.Plan, tail,
				int(avg), value.MaxMilli, value.Count, value.Namespace, red, str, tail)
		} else {
			output = fmt.Sprintf("|%-10s %8s %6d %8d %6d %-33s %-62s|\n", value.Command, "",
				int(avg), value.MaxMilli, value.Count, value.Namespace, str)
		}
		buffer.WriteString(output)
		if len(value.Filter) > 60 {
			remaining := value.Filter[len(str):]
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
				if value.Plan == COLLSCAN {
					output = fmt.Sprintf("|%74s   %v%-62s%v|\n", " ", red, pstr, tail)
					buffer.WriteString(output)
				} else {
					output = fmt.Sprintf("|%74s   %-62s|\n", " ", pstr)
					buffer.WriteString(output)
				}
			}
		}
		if value.Index != "" {
			output = fmt.Sprintf("|...index:  %v%-128s%v|\n", green, value.Index, tail)
			buffer.WriteString(output)
		}
	}
	buffer.WriteString("+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	summaries = append(summaries, buffer.String())
	if maxSize < count {
		summaries = append(summaries, fmt.Sprintf(`top %d of %d lines displayed.`, maxSize, count))
	}
	return strings.Join(summaries, "\n")
}
