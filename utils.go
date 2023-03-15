/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * utils.go
 */

package hatchet

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	MAX_SIZE  = 128
	TAIL_SIZE = 7
)

// ToInt converts to int
func ToInt(num interface{}) int {
	f := fmt.Sprintf("%v", num)
	x, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return 0
	}
	return int(x)
}

func replaceSpecialChars(name string) string {
	for _, sep := range []string{"-", ".", " ", ":", ","} {
		name = strings.ReplaceAll(name, sep, "_")
	}
	return name
}

func getHatchetName(filename string) string {
	temp := filepath.Base(filename)
	hatchetName := replaceSpecialChars(temp)
	i := strings.LastIndex(hatchetName, "_log")
	if i >= 0 && i >= len(temp)-len(".log.gz") {
		hatchetName = replaceSpecialChars(hatchetName[0:i])
	}
	if len(hatchetName) > MAX_SIZE {
		hatchetName = hatchetName[:MAX_SIZE-TAIL_SIZE]
	}
	if i = strings.LastIndex(hatchetName, "_gz"); i > 0 {
		hatchetName = hatchetName[:i]
	}
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, TAIL_SIZE)
	rand.Read(b)
	tail := fmt.Sprintf("%x", b)[:TAIL_SIZE-1]

	r := []rune(hatchetName) // convert string to runes
	if unicode.IsDigit(r[0]) {
		hatchetName = "_" + hatchetName
	}
	return fmt.Sprintf("%v_%v", hatchetName, tail)
}

func EscapeString(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z"}
	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}
	return value
}

func GetDateSubString(start string, end string) string {
	var err error
	substr := "SUBSTR(date, 1, 16)"
	if len(start) < 16 || len(end) < 16 {
		return substr
	}
	var stime, etime time.Time
	layout := "2006-01-02T15:04"
	if stime, err = time.Parse(layout, start[:16]); err != nil {
		return substr
	}
	if etime, err = time.Parse(layout, end[:16]); err != nil {
		return substr
	}
	minutes := etime.Sub(stime).Minutes()
	if minutes < 1 {
		return "SUBSTR(date, 1, 19)"
	} else if minutes < 10 {
		return "SUBSTR(date, 1, 18)||'9'"
	} else if minutes < 60 {
		return "SUBSTR(date, 1, 16)||':59'"
	} else if minutes < 600 {
		return "SUBSTR(date, 1, 15)||'9:59'"
	} else if minutes < 3600 {
		return "SUBSTR(date, 1, 13)||':59:59'"
	} else {
		return "SUBSTR(date, 1, 10)||'T23:59:59'"
	}
}

func GetHatchetSummary(info HatchetInfo) string {
	arr := []string{}
	if info.Module == "" {
		info.Module = "community"
	}
	if info.Version != "" {
		arr = append(arr, fmt.Sprintf(": MongoDB v%v (%v)", info.Version, info.Module))
	}
	if info.OS != "" {
		arr = append(arr, "os: "+info.OS)
	}
	if info.Arch != "" {
		arr = append(arr, "arch: "+info.Arch)
	}
	return info.Name + strings.Join(arr, ", ")
}

// GetOffsetLimit returns offset, limit
func GetOffsetLimit(str string) (int, int) {
	toks := strings.Split(str, ",")
	if len(toks) >= 2 {
		return ToInt(toks[0]), ToInt(toks[1])
	} else if len(toks) == 1 {
		return 0, ToInt(toks[0])
	}
	return 0, 0
}

func getDateTimeStr(tm time.Time) string {
	dt := tm.Format("2006-01-02T15:04:05.000-0000")
	return dt
}

func GetBufioReader(data []byte) (*bufio.Reader, error) {
	isGzipped := false
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		isGzipped = true
	}

	if isGzipped {
		gzipReader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()

		var buf bytes.Buffer
		if _, err = buf.ReadFrom(gzipReader); err != nil {
			return nil, err
		}

		return bufio.NewReader(&buf), nil
	}

	return bufio.NewReader(bytes.NewReader(data)), nil
}
