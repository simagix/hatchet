// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"database/sql"
	"testing"
)

func TestGetSubStringFromTable(t *testing.T) {
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	instr := `CREATE TABLE IF NOT EXISTS hatchet ( name text not null primary key,
				version text, module text, arch text, os text, start text, end text);
			  INSERT INTO hatchet (name, version, module, arch, os, start, end)
				VALUES ('hatchet_test', '', '', '', '', '2023-01-01T12:11:02Z', '2023-01-10T12:34:20Z');`
	if _, err = db.Exec(instr); err != nil {
		t.Fatal(err)
	}
	value := getSubStringFromTable("hatchet_test")
	if value != "SUBSTR(date, 1, 13)||':59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 13)||':59'", "but got", value)
	}
	t.Log(value)
	if _, err = db.Exec("DELETE FROM hatchet WHERE name = 'hatchet_test';"); err != nil {
		t.Fatal(err)
	}
}

func TestGetDateSubString(t *testing.T) {
	value := getDateSubString("2023-01-01T12:11:02Z", "2023-01-01T14:35:20Z")
	if value != "SUBSTR(date, 1, 16)||':59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 16)||':59'", "but got", value)
	}
	t.Log(value)

	value = getDateSubString("2023-01-01T12:11:02Z", "2023-01-02T12:34:20Z")
	if value != "SUBSTR(date, 1, 15)||'9:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 15)||'9:59'", "but got", value)
	}
	t.Log(value)

	value = getDateSubString("2023-01-01T12:11:02Z", "2023-02-10T12:34:20Z")
	if value != "SUBSTR(date, 1, 10)||'T23:59:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 13)||':59:59'", "but got", value)
	}
	t.Log(value)
}
