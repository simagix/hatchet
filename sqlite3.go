// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"database/sql"
	"fmt"
	"log"
)

type SQLite3DB struct {
	clientStmt *sql.Stmt
	db         *sql.DB
	dbfile     string
	tableName  string
	tx         *sql.Tx
	pstmt      *sql.Stmt
}

func NewSQLite3DB(dbfile string, tableName string) (*SQLite3DB, error) {
	var err error
	logv2db := &SQLite3DB{dbfile: dbfile, tableName: tableName}
	if logv2db.db, err = sql.Open("sqlite3", dbfile); err != nil {
		return logv2db, err
	}
	return logv2db, err
}

func (ptr *SQLite3DB) Begin() error {
	var err error
	log.Println("creating table", ptr.tableName)
	stmts := getHatchetInitStmt(ptr.tableName)
	if _, err = ptr.db.Exec(stmts); err != nil {
		return err
	}
	if ptr.tx, err = ptr.db.Begin(); err != nil {
		return err
	}
	if ptr.pstmt, err = ptr.tx.Prepare(getHatchetPreparedStmt(ptr.tableName)); err != nil {
		return err
	}
	if ptr.clientStmt, err = ptr.tx.Prepare(getClientPreparedStmt(ptr.tableName)); err != nil {
		return err
	}
	return err
}

func (ptr *SQLite3DB) Commit() error {
	return ptr.tx.Commit()
}

func (ptr *SQLite3DB) Close() error {
	err := ptr.pstmt.Close()
	if err != nil {
		return err
	}
	return ptr.clientStmt.Close()
}

func (ptr *SQLite3DB) InsertLog(index int, end string, doc *Logv2Info, stat *OpStat) error {
	var err error
	_, err = ptr.pstmt.Exec(index, end, doc.Severity, doc.Component, doc.Context,
		doc.Msg, doc.Attributes.PlanSummary, doc.Attr.Map()["type"], doc.Attributes.NS, doc.Message,
		stat.Op, stat.QueryPattern, stat.Index, doc.Attributes.Milli, doc.Attributes.Reslen)
	return err
}

func (ptr *SQLite3DB) InsertClientConn(index int, rmt Remote) error {
	var err error
	_, err = ptr.clientStmt.Exec(index, rmt.Value, rmt.Port, rmt.Conns, rmt.Accepted, rmt.Ended)
	return err
}

func (ptr *SQLite3DB) UpdateHatchetInfo(info HatchetInfo) error {
	istmt := fmt.Sprintf(`INSERT OR REPLACE INTO hatchet (name, version, module, arch, os, start, end)
		VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v');`, ptr.tableName, info.Version, info.Module, info.Arch, info.OS, info.Start, info.End)
	_, err := ptr.db.Exec(istmt)
	return err
}

func (ptr *SQLite3DB) InsertOps() error {
	istmt := fmt.Sprintf(`INSERT INTO %v_ops
			SELECT op, COUNT(*), ROUND(AVG(milli),1), MAX(milli), SUM(milli), ns, _index, SUM(reslen), filter
				FROM %v WHERE op != "" GROUP BY op, ns, filter, _index`, ptr.tableName, ptr.tableName)
	_, err := ptr.db.Exec(istmt)
	return err
}
