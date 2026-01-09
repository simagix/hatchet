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
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	cacheSize   int
	from        time.Time
	logname     string
	legacy      bool
	hatchetName string
	isDigest    bool
	merge       bool
	s3client    *S3Client
	testing     bool //test mode
	to          time.Time
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
	Marker     int
}

type Attributes struct {
	AppName            string                 `json:"appName" bson:"appName"`
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

	Marker int
}

type LegacyLog struct {
	Timestamp string `json:"date" bson:"date"`
	Severity  string `json:"severity" bson:"severity"`
	Component string `json:"component" bson:"component"`
	Context   string `json:"context" bson:"context"`
	Marker    int
	Message   string `json:"message" bson:"message"` // remaining legacy message
}

type HatchetInfo struct {
	Arch    string `bson:"arch"`
	End     string `bson:"end"`
	Merge   bool   `bson:"merge"`
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

// isMongoDBLog checks if a file is a MongoDB log by reading and validating the first line
func isMongoDBLog(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	reader, err := gox.NewReader(file)
	if err != nil {
		return false
	}

	// Read first non-empty line
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return false
		}
		if len(line) == 0 {
			continue
		}
		// Try to parse as MongoDB logv2 JSON
		doc := Logv2Info{}
		if err := bson.UnmarshalExtJSON(line, false, &doc); err != nil {
			return false
		}
		// Check for required logv2 fields: timestamp and severity
		return !doc.Timestamp.IsZero() && doc.Severity != ""
	}
}

// Analyze analyzes logs from a file or directory (1 level only)
func (ptr *Logv2) Analyze(logname string, marker int) error {
	// Check if input is a directory
	fileInfo, err := os.Stat(logname)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		// Process directory (1 level only, no recursion)
		entries, err := os.ReadDir(logname)
		if err != nil {
			return err
		}
		// Get existing names once, then track new names locally to avoid DB query timing issues
		existingNames, _ := GetExistingHatchetNames()
		processedNames := make([]string, 0)
		fileCount := 0
		for _, entry := range entries {
			if entry.IsDir() {
				continue // skip subdirectories
			}
			name := entry.Name()
			if strings.HasPrefix(name, ".") {
				continue // skip hidden files
			}
			fullPath := filepath.Join(logname, name)
			// Check if file is a MongoDB log by reading first line
			if !isMongoDBLog(fullPath) {
				log.Printf("skipping %s (not a MongoDB log)", name)
				continue
			}
			fileCount++
			// Generate unique name using combined existing + processed names
			allNames := append(existingNames, processedNames...)
			hatchetName := getUniqueHatchetName(fullPath, allNames)
			processedNames = append(processedNames, hatchetName)
			// Create a new Logv2 instance for each file to avoid state issues
			fileLogv2 := &Logv2{
				hatchetName: hatchetName,
				url:         ptr.url,
				legacy:      ptr.legacy,
				from:        ptr.from,
				to:          ptr.to,
			}
			if err := fileLogv2.Analyze(fullPath, 0); err != nil { // marker=0 to skip name regeneration
				log.Printf("error processing %s: %v", fullPath, err)
				// continue with other files
			}
			// Print summary for each file
			if !fileLogv2.legacy {
				fileLogv2.PrintSummary()
			}
		}
		if fileCount == 0 {
			log.Printf("no MongoDB log files found in directory %s", logname)
		}
		// Clear hatchetName so caller knows not to print summary again
		ptr.hatchetName = ""
		return nil
	}

	var buf []byte
	var file *os.File
	var reader *bufio.Reader
	ptr.logname = logname
	// Generate unique hatchet name for each file when not merging
	// marker=0: name pre-set by directory handler or upload handler, skip regeneration
	// marker>=1: command line files, generate name for each
	if !ptr.merge && marker > 0 {
		existingNames, _ := GetExistingHatchetNames()
		ptr.hatchetName = getUniqueHatchetName(ptr.logname, existingNames)
	}
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
			// EOF from NewReader is expected for empty files, so don't treat it as an error
			if errors.Is(err, io.EOF) {
				// For empty files, create a simple reader and continue
				if _, err = file.Seek(0, 0); err != nil {
					return err
				}
				reader = bufio.NewReader(file)
				ptr.totalLines = 0
			} else {
				return err
			}
		} else if !ptr.legacy {
			log.Println("fast counting", logname, "...")
			ptr.totalLines, _ = gox.CountLines(reader)
			log.Println("counted", ptr.totalLines, "lines")
			if _, err = file.Seek(0, 0); err != nil {
				return err
			}
			if reader, err = gox.NewReader(file); err != nil {
				// EOF from NewReader is expected for empty files
				if errors.Is(err, io.EOF) {
					if _, err = file.Seek(0, 0); err != nil {
						return err
					}
					reader = bufio.NewReader(file)
				} else {
					return err
				}
			}
		}
	}

	var isPrefix bool
	index := 0
	var start, end string
	var dbase Database
	var mu sync.Mutex   // protects buildInfo, start, and end
	var dbMu sync.Mutex // protects database operations (SQLite prepared statements aren't thread-safe)

	if !ptr.legacy {
		if dbase, err = GetDatabase(ptr.hatchetName); err != nil {
			return err
		}
		defer dbase.Close()
		if err = dbase.Begin(); err != nil {
			return err
		}
	}

	threads := runtime.NumCPU() - 1
	if threads == 0 {
		threads = 1
	}
	log.Printf("using %v threads\n", threads)
	failedMap := FailedMessages{counters: map[string]int{}}
	var wg = gox.NewWaitGroup(threads)
	var lineBuf bytes.Buffer // Reusable buffer for multi-line entries
	for {
		if !ptr.testing && !ptr.legacy && index%50 == 0 && ptr.totalLines > 0 {
			fmt.Fprintf(os.Stderr, "\r%3d%% \r", (100*index)/ptr.totalLines)
		}
		if buf, isPrefix, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
			break
		}
		index++
		if len(buf) == 0 {
			log.Println("line", index, "is blank.")
			continue
		}
		var str string
		if !isPrefix {
			// Fast path: single line, no allocation needed beyond the string conversion
			str = string(buf)
		} else {
			// Multi-line entry: use buffer to avoid repeated string concatenation
			lineBuf.Reset()
			lineBuf.Write(buf)
			for isPrefix {
				var bbuf []byte
				if bbuf, isPrefix, err = reader.ReadLine(); err != nil {
					// EOF in the inner loop means incomplete line, which is an error
					if errors.Is(err, io.EOF) {
						err = fmt.Errorf("unexpected EOF while reading multi-line prefix")
					}
					break
				}
				lineBuf.Write(bbuf)
			}
			str = lineBuf.String()
		}
		// If we got an error in the inner loop, break from outer loop
		if err != nil {
			break
		}

		wg.Add(1)
		go func(index int, instr string, marker int) {
			defer wg.Done()
			var localErr error // Use local error variable to avoid race condition
			doc := Logv2Info{}
			if localErr = bson.UnmarshalExtJSON([]byte(instr), false, &doc); localErr != nil {
				log.Println("error UnmarshalExtJSON line", index, localErr)
				return
			}
			doc.Marker = marker
			if localErr = SetRawJSONMessage(&doc, instr); localErr != nil {
				log.Println("error SetRawJSONMessage line", index, localErr)
				return
			}
			// Protect buildInfo access with mutex
			mu.Lock()
			if ptr.buildInfo == nil && doc.Msg == "Build Info" {
				attrMap := BsonD2M(doc.Attr)
				ptr.buildInfo = attrMap["buildInfo"].(bson.M)
			}
			buildInfoCopy := ptr.buildInfo
			mu.Unlock()
			if buildInfoCopy != nil && (doc.Timestamp.Before(ptr.from) || doc.Timestamp.After(ptr.to)) {
				return
			}
			if ptr.legacy {
				dt := getDateTimeStr(doc.Timestamp)
				logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", dt,
					doc.Severity, doc.Component, doc.Context, doc.Message)
				if !ptr.testing {
					fmt.Println(logstr)
				}
				return
			}
			stat, _ := AnalyzeSlowOp(&doc)
			docEnd := getDateTimeStr(doc.Timestamp)
			// Protect start and end access with mutex
			mu.Lock()
			if start == "" {
				start = docEnd
			}
			end = docEnd
			mu.Unlock()

			// Serialize database operations - SQLite prepared statements aren't thread-safe
			if localErr = insertLogData(dbase, &dbMu, index, docEnd, &doc, stat); localErr != nil {
				log.Println("error inserting log data line", index, localErr)
				return
			}
			failed := " failed"
			if strings.Contains(doc.Message, failed) {
				n := strings.Index(doc.Message, failed) + len(failed)
				failedMap.inc(doc.Message[:n])
			}
		}(index, str, marker)
	}
	// EOF is expected when we've finished reading the file, so don't treat it as an error
	if errors.Is(err, io.EOF) {
		err = nil
	} else if err != nil {
		return err
	}
	wg.Wait()
	log.Println("completed parsing logs")
	if !ptr.testing && !ptr.legacy {
		fmt.Fprintf(os.Stderr, "\r                         \r")
	}
	if ptr.legacy {
		return nil
	}
	if err = dbase.Commit(); err != nil {
		log.Println("error commit", err)
		return err
	}
	if err = dbase.InsertFailedMessages(&failedMap); err != nil {
		log.Println("error insert failed messages", err)
		return err
	}
	info := HatchetInfo{Start: start, End: end, Merge: ptr.merge}
	if ptr.buildInfo != nil {
		if ptr.buildInfo["environment"] != nil {
			env := ptr.buildInfo["environment"].(bson.M)
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
		log.Println("error update Hatchet info", err)
		return err
	}
	return nil
}

// insertLogData serializes database operations with a mutex.
// SQLite prepared statements within a transaction aren't thread-safe.
func insertLogData(dbase Database, dbMu *sync.Mutex, index int, docEnd string, doc *Logv2Info, stat *OpStat) error {
	dbMu.Lock()
	defer dbMu.Unlock()

	if err := dbase.InsertLog(index, docEnd, doc, stat); err != nil {
		return err
	}
	if doc.Client != nil {
		if (doc.Client.Accepted + doc.Client.Ended) > 0 { // record connections
			if err := dbase.InsertClientConn(index, doc); err != nil {
				return err
			}
		} else if doc.Client.Driver != "" {
			if isAppDriver(doc.Client) {
				if err := dbase.InsertDriver(index, doc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ptr *Logv2) PrintSummary() error {
	// Skip if no hatchet name (e.g., after directory processing)
	if ptr.hatchetName == "" {
		return nil
	}
	dbase, err := GetDatabase(ptr.hatchetName)
	if err != nil {
		return err
	}
	if err = dbase.CreateMetaData(); err != nil {
		log.Println("error create metadata", err)
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

	if strings.HasPrefix(driver, "NetworkInterfaceTL") || driver == "MongoDB Internal Client" {
		return false
	} else if driver == "mongo-go-driver" && strings.HasSuffix(version, "-cloud") {
		return false
	}
	return true
}
