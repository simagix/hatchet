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
	verbose    bool
}

func NewSQLite3DB(dbfile string, tableName string) (*SQLite3DB, error) {
	var err error
	sqlite := &SQLite3DB{dbfile: dbfile, tableName: tableName}
	if sqlite.db, err = sql.Open("sqlite3", dbfile); err != nil {
		return sqlite, err
	}
	return sqlite, err
}

func (ptr *SQLite3DB) GetVerbose() bool {
	return ptr.verbose
}

func (ptr *SQLite3DB) SetVerbose(b bool) {
	ptr.verbose = b
}

func (ptr *SQLite3DB) Begin() error {
	var err error
	log.Println("creating table", ptr.tableName)
	stmts := ptr.GetHatchetInitStmt(ptr.tableName)
	if _, err = ptr.db.Exec(stmts); err != nil {
		return err
	}
	if ptr.tx, err = ptr.db.Begin(); err != nil {
		return err
	}
	if ptr.pstmt, err = ptr.tx.Prepare(ptr.GetHatchetPreparedStmt(ptr.tableName)); err != nil {
		return err
	}
	if ptr.clientStmt, err = ptr.tx.Prepare(ptr.GetClientPreparedStmt(ptr.tableName)); err != nil {
		return err
	}
	return err
}

func (ptr *SQLite3DB) Commit() error {
	return ptr.tx.Commit()
}

func (ptr *SQLite3DB) Close() error {
	var err error
	if ptr.pstmt != nil {
		if err = ptr.pstmt.Close(); err != nil {
			return err
		}
	}
	if ptr.clientStmt != nil {
		if err = ptr.clientStmt.Close(); err != nil {
			return err
		}
	}
	defer ptr.db.Close()
	return err
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

// GetHatchetInitStmt returns init statement
func (ptr *SQLite3DB) GetHatchetInitStmt(tableName string) string {
	return fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS hatchet ( name text not null primary key,
				version text, module text, arch text, os text, start text, end text);

			DROP TABLE IF EXISTS %v;
			CREATE TABLE %v (
				id integer not null primary key, date text, severity text, component text, context text,
				msg text, plan text, type text, ns text, message text,
				op text, filter text, _index text, milli integer, reslen integer);

			DROP TABLE IF EXISTS %v_ops;
			CREATE TABLE %v_ops ( op, count, avg_ms, max_ms, total_ms, ns, _index, reslen, filter);

			CREATE INDEX IF NOT EXISTS %v_idx_component ON %v (component);
			CREATE INDEX IF NOT EXISTS %v_idx_context ON %v (context);
			CREATE INDEX IF NOT EXISTS %v_idx_severity ON %v (severity);
			CREATE INDEX IF NOT EXISTS %v_idx_op ON %v (op,ns,filter);

			DROP TABLE IF EXISTS %v_rmt;
			CREATE TABLE %v_rmt(
				id integer not null primary key, ip text, port text, conns integer, accepted integer, ended integer)`,
		tableName, tableName, tableName, tableName, tableName,
		tableName, tableName, tableName, tableName, tableName,
		tableName, tableName, tableName, tableName)
}

// GetHatchetPreparedStmt returns prepared statement of log table
func (ptr *SQLite3DB) GetHatchetPreparedStmt(id string) string {
	return fmt.Sprintf(`INSERT INTO %v (id, date, severity, component, context,
		msg, plan, type, ns, message, op, filter, _index, milli, reslen)
		VALUES(?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?)`, id)
}

// GetClientPreparedStmt returns prepared statement of client table
func (ptr *SQLite3DB) GetClientPreparedStmt(id string) string {
	return fmt.Sprintf(`INSERT INTO %v_rmt (id, ip, port, conns, accepted, ended)
		VALUES(?,?,?,?,?, ?)`, id)
}
