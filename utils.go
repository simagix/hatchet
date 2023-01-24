// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	tableName := replaceSpecialChars(temp)
	i := strings.LastIndex(tableName, "_log")
	if i >= 0 && i >= len(temp)-len(".log.gz") {
		tableName = replaceSpecialChars(tableName[0:i])
	}
	if len(tableName) > MAX_SIZE {
		tableName = tableName[:MAX_SIZE-TAIL_SIZE]
	}
	if i = strings.LastIndex(tableName, "_gz"); i > 0 {
		tableName = tableName[:i]
	}
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, TAIL_SIZE)
	rand.Read(b)
	tail := fmt.Sprintf("%x", b)[:TAIL_SIZE-1]
	return fmt.Sprintf("%v_%v", tableName, tail)
}

func EscapeString(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z"}
	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}
	return value
}
