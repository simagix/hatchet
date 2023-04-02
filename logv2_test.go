/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * logv2_test.go
 */

package hatchet

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/mattn/go-sqlite3"
)

func TestAnalyze(t *testing.T) {
	regex := func(re, s string) (bool, error) {
		return regexp.MatchString(re, s)
	}
	sql.Register("sqlite3_extended",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regex, true)
			},
		})
	filename := "testdata/mongod_ops.log.gz"
	logv2 := &Logv2{testing: true, url: SQLITE3_FILE}
	err := logv2.Analyze(filename)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzeLegacy(t *testing.T) {
	filename := "testdata/mongod_ops.log.gz"
	logv2 := &Logv2{testing: true, legacy: true}
	err := logv2.Analyze(filename)
	if err != nil {
		t.Fatal(err)
	}
}
