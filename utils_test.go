// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestToInt(t *testing.T) {
	input := "23"
	value := ToInt(input)
	if value != 23 {
		t.Fatal("expected", 23, "but got", value)
	}

	str := ""
	value = ToInt(str)
	if value != 0 {
		t.Fatal("expected", 0, "but got", value)
	}
}

func TestReplaceSpecialChars(t *testing.T) {
	value := "a_b_c_d_e_name"

	filename := "a-b.c d:e,name"
	fname := replaceSpecialChars(filename)
	if value != fname {
		t.Fatal("expected", value, "but got", fname)
	}
}

func TestGetHatchetName(t *testing.T) {
	filename := "mongod.log"
	length := len(filename) - len(".log") + TAIL_SIZE
	hatchetName := getHatchetName(filename)
	if len(hatchetName) != length {
		t.Fatal("expected", length, "but got", len(hatchetName))
	}

	filename = "mongod.log.gz"
	length = len(filename) - len(".log.gz") + TAIL_SIZE
	hatchetName = getHatchetName(filename)
	if len(hatchetName) != length {
		t.Fatal("expected", length, "but got", len(hatchetName))
	}

	filename = "mongod"
	length = len(filename) + TAIL_SIZE
	hatchetName = getHatchetName(filename)
	if len(hatchetName) != length {
		t.Fatal("expected", length, "but got", len(hatchetName))
	}

	filename = "filesys-shard-00-01.abcde.mongodb.net_2021-07-24T10_12_58_2021-07-25T10_12_58_mongodb.log.gz"
	length = len(filename) + TAIL_SIZE
	hatchetName = getHatchetName(filename)
	modified := "filesys_shard_00_01_abcde_mongodb_net_2021_07_24T10_12_58_2021_07_25T10_12_58_mongodb"
	if !strings.HasPrefix(hatchetName, modified) {
		t.Fatal(modified+"_*", length, "but got", hatchetName)
	}

	filename = "testdata/demo_errmsg.log.gz"
	fname := filepath.Base((filename))
	t.Log(fname)
	hatchetName = getHatchetName(filename)
	t.Log(hatchetName)
	if len(hatchetName) != len(fname) {
		t.Fatal("expected", len(fname), "but got", len(hatchetName))
	}

	filename = "testdata/0_errmsg.log.gz"
	fname = filepath.Base((filename))
	t.Log(fname)
	hatchetName = getHatchetName(filename)
	t.Log(hatchetName)
	if len(hatchetName) != len(fname)+1 { // added _
		t.Fatal("expected", len(fname), "but got", len(hatchetName))
	}
}

func TestGetDateSubString(t *testing.T) {
	value := GetDateSubString("2023-01-01T12:11:02Z", "2023-01-01T14:35:20Z")
	if value != "SUBSTR(date, 1, 15)||'9:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 15)||'9:59'", "but got", value)
	}
	t.Log(value)

	value = GetDateSubString("2023-01-01T12:11:02Z", "2023-01-02T12:34:20Z")
	if value != "SUBSTR(date, 1, 13)||':59:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 13)||':59:59'", "but got", value)
	}
	t.Log(value)

	value = GetDateSubString("2023-01-01T12:11:02Z", "2023-02-10T12:34:20Z")
	if value != "SUBSTR(date, 1, 10)||'T23:59:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 13)||':59:59'", "but got", value)
	}
	t.Log(value)
}

func TestGetOffsetLimit(t *testing.T) {
	limit := "100"
	o, l := GetOffsetLimit(limit)
	if o != 0 || l != 100 {
		t.Fatal("expected", 0, 100, "but got", o, l)
	}

	limit = "100,100"
	o, l = GetOffsetLimit(limit)
	if o != 100 || l != 100 {
		t.Fatal("expected", 100, 100, "but got", o, l)
	}
}
