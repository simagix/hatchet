// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"database/sql"
	"testing"
)

func TestGetSubStringFromTable(t *testing.T) {
	var dbase Database
	db, err := sql.Open("sqlite3", "data/hatchet.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec("DELETE FROM hatchet WHERE name = 'hatchet_test';"); err != nil {
		t.Fatal(err)
	}
	instr := `CREATE TABLE IF NOT EXISTS hatchet ( name text not null primary key,
				version text, module text, arch text, os text, start text, end text);
			  INSERT INTO hatchet (name, version, module, arch, os, start, end)
				VALUES ('hatchet_test', '', '', '', '', '2023-01-01T12:11:02Z', '2023-01-10T12:34:20Z');`
	if _, err = db.Exec(instr); err != nil {
		t.Fatal(err)
	}
	if dbase, err = NewSQLite3DB("data/hatchet.db", "hatchet_test"); err != nil {
		t.Fatal(err)
	}
	value := dbase.GetSubStringFromTable("hatchet_test")
	if value != "SUBSTR(date, 1, 10)||'T23:59:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 10)||'T23:59:59'", "but got", value)
	}
	t.Log(value)
	if _, err = db.Exec("DELETE FROM hatchet WHERE name = 'hatchet_test';"); err != nil {
		t.Fatal(err)
	}
}
