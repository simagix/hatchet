// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"testing"
)

func TestAnalyze(t *testing.T) {
	filename := "testdata/mongod_ops.log.gz"
	logv2 := &Logv2{testing: true, dbfile: SQLITE3_FILE}
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
