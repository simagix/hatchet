/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * sqlite3.go
 */

package hatchet

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type SQLite3DB struct {
	clientStmt  *sql.Stmt // {hatchet}_clients
	driverStmt  *sql.Stmt // {hatchet}_drivers
	db          *sql.DB
	dbfile      string
	hatchetName string
	tx          *sql.Tx
	pstmt       *sql.Stmt // {hatchet}
	verbose     bool
}

func NewSQLite3DB(dbfile string, hatchetName string, cacheSize int) (*SQLite3DB, error) {
	var err error
	sqlite := &SQLite3DB{dbfile: dbfile, hatchetName: hatchetName}
	dirname := filepath.Dir(dbfile)
	os.Mkdir(dirname, 0755)
	if sqlite.db, err = sql.Open("sqlite3_extended", dbfile); err != nil {
		return nil, err
	}
	if cacheSize > 0 && cacheSize != 2000 {
		pragma := fmt.Sprintf("PRAGMA cache_size = %d;", cacheSize)
		if _, err = sqlite.db.Exec(pragma); err != nil {
			sqlite.db.Close()
			return nil, err
		}
	}
	return sqlite, nil
}

func (ptr *SQLite3DB) GetVerbose() bool {
	return ptr.verbose
}

func (ptr *SQLite3DB) SetVerbose(b bool) {
	ptr.verbose = b
}

func (ptr *SQLite3DB) Begin() error {
	var err error
	log.Println("creating hatchet", ptr.hatchetName)
	if err = CreateTables(ptr.db, ptr.hatchetName); err != nil {
		return err
	}
	if ptr.tx, err = ptr.db.Begin(); err != nil {
		return err
	}
	if ptr.pstmt, err = ptr.tx.Prepare(GetHatchetPreparedStmt(ptr.hatchetName)); err != nil {
		return err
	}
	if ptr.clientStmt, err = ptr.tx.Prepare(GetClientPreparedStmt(ptr.hatchetName)); err != nil {
		return err
	}
	if ptr.driverStmt, err = ptr.tx.Prepare(GetDriverPreparedStmt(ptr.hatchetName)); err != nil {
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
	if ptr.driverStmt != nil {
		if err = ptr.driverStmt.Close(); err != nil {
			return err
		}
	}
	defer ptr.db.Close()
	return err
}

// Drop drops all tables of a hatchet
func (ptr *SQLite3DB) Drop() error {
	var err error
	hatchetName := ptr.hatchetName
	stmts := fmt.Sprintf(`
			DROP TABLE IF EXISTS %v;
			DROP TABLE IF EXISTS %v_audit;
			DROP TABLE IF EXISTS %v_clients;
			DROP TABLE IF EXISTS %v_drivers;
			DROP TABLE IF EXISTS %v_ops;

			DROP INDEX IF EXISTS %v_idx_component_severity;
			DROP INDEX IF EXISTS %v_idx_context_date;
			DROP INDEX IF EXISTS %v_idx_context_op_reslen;
			DROP INDEX IF EXISTS %v_idx_ns_reslen;
			DROP INDEX IF EXISTS %v_idx_ns_op_filter;
			DROP INDEX IF EXISTS %v_idx_milli;
			DROP INDEX IF EXISTS %v_idx_severity;

			DROP INDEX IF EXISTS %v_audit_idx_type_value;
			DROP INDEX IF EXISTS %v_clients_idx_ip_accepted;
			DROP INDEX IF EXISTS %v_clients_idx_ip_context;
			DROP INDEX IF EXISTS %v_drivers_idx_driver_version;
			DROP INDEX IF EXISTS %v_ops_idx_avgms;
			DROP INDEX IF EXISTS %v_ops_idx_index;`,
		hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName,
		hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName,
	)
	if _, err = ptr.db.Exec(stmts); err != nil {
		return err
	}

	stmt := fmt.Sprintf(`DELETE FROM hatchet WHERE name = '%v'`, ptr.hatchetName)
	if _, err := ptr.db.Exec(stmt); err != nil {
		return err
	}
	return err
}

func (ptr *SQLite3DB) InsertLog(index int, end string, doc *Logv2Info, stat *OpStat) error {
	var err error
	_, err = ptr.pstmt.Exec(index, end, doc.Severity, doc.Component, doc.Context,
		doc.Msg, doc.Attributes.PlanSummary, doc.Attr.Map()["type"], doc.Attributes.NS, doc.Message,
		stat.Op, stat.QueryPattern, stat.Index, doc.Attributes.Milli, doc.Attributes.Reslen)
	return err
}

func (ptr *SQLite3DB) InsertClientConn(index int, doc *Logv2Info) error {
	var err error
	client := doc.Client
	_, err = ptr.clientStmt.Exec(index, client.IP, client.Port, client.Conns, client.Accepted, client.Ended, doc.Context)
	return err
}

func (ptr *SQLite3DB) InsertDriver(index int, doc *Logv2Info) error {
	var err error
	client := doc.Client
	_, err = ptr.driverStmt.Exec(index, client.IP, client.Driver, client.Version)
	return err
}

func (ptr *SQLite3DB) InsertFailedMessages(m *FailedMessages) error {
	var err error
	for k, v := range m.counters {
		stmt := fmt.Sprintf("INSERT INTO %v_audit (type, name, value) VALUES ('failed','%s', %d)", ptr.hatchetName, k, v)
		if _, err = ptr.db.Exec(stmt); err != nil {
			return err
		}
	}
	return err
}

func (ptr *SQLite3DB) UpdateHatchetInfo(info HatchetInfo) error {
	istmt := fmt.Sprintf(`INSERT OR REPLACE INTO hatchet (name, version, module, arch, os, start, end)
		VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v');`, ptr.hatchetName, info.Version, info.Module, info.Arch, info.OS, info.Start, info.End)
	_, err := ptr.db.Exec(istmt)
	return err
}

func (ptr *SQLite3DB) CreateMetaData() error {
	var err error
	log.Println("creating indexes and this may take minutes")
	if err = CreateIndexes(ptr.db, ptr.hatchetName); err != nil {
		return err
	}
	log.Printf("insert ops into %v_ops\n", ptr.hatchetName)
	istmt := fmt.Sprintf(`INSERT INTO %v_ops
			SELECT op, COUNT(*), ROUND(AVG(milli),1), MAX(milli), SUM(milli), ns, _index, SUM(reslen), filter
				FROM %v WHERE op != "" GROUP BY op, ns, filter, _index`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		log.Println(istmt)
		explain(ptr.db, istmt)
	}
	if _, err = ptr.db.Exec(istmt); err != nil {
		return err
	}

	log.Printf("insert [exception] into %v_audit\n", ptr.hatchetName)
	istmt = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'exception', severity, COUNT(*) count FROM %v WHERE severity IN ('W', 'E', 'F') 
		GROUP by severity`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		log.Println(istmt)
		explain(ptr.db, istmt)
	}
	if _, err = ptr.db.Exec(istmt); err != nil {
		return err
	}

	log.Printf("insert [op] into %v_audit\n", ptr.hatchetName)
	istmt = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'op', op, COUNT(*) count FROM %v WHERE op != '' GROUP by op`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		log.Println(istmt)
		explain(ptr.db, istmt)
	}
	if _, err = ptr.db.Exec(istmt); err != nil {
		return err
	}

	log.Printf("insert [ip] into %v_audit\n", ptr.hatchetName)
	istmt = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'ip', ip, SUM(accepted) open FROM %v_clients GROUP by ip`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		log.Println(istmt)
		explain(ptr.db, istmt)
	}
	if _, err = ptr.db.Exec(istmt); err != nil {
		return err
	}

	log.Printf("insert [ns] into %v_audit\n", ptr.hatchetName)
	istmt = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'ns', ns, COUNT(*) count FROM %v WHERE op != "" GROUP by ns`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		log.Println(istmt)
		explain(ptr.db, istmt)
	}
	if _, err = ptr.db.Exec(istmt); err != nil {
		return err
	}

	log.Printf("insert [reslen-ns] into %v_audit\n", ptr.hatchetName)
	istmt = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'reslen-ns', ns, SUM(reslen) FROM %v WHERE ns != "" AND reslen > 0 GROUP by ns`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		log.Println(istmt)
		explain(ptr.db, istmt)
	}
	if _, err = ptr.db.Exec(istmt); err != nil {
		return err
	}

	log.Printf("insert [reslen-ip] into %v_audit\n", ptr.hatchetName)
	istmt = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'reslen-ip', ip, SUM(reslen) FROM (
			SELECT a.context, sum(reslen) reslen, b.ip ip FROM %v a, %v_clients b
				WHERE op != "" and reslen > 0 and a.context = b.context GROUP by a.context
		) GROUP BY ip`,
		ptr.hatchetName, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		log.Println(istmt)
		explain(ptr.db, istmt)
	}
	if _, err = ptr.db.Exec(istmt); err != nil {
		return err
	}
	return err
}

// CreateTables returns init statement
func CreateTables(db *sql.DB, hatchetName string) error {
	var err error
	tables := []string{
		`CREATE TABLE IF NOT EXISTS hatchet (
			name text not null primary key,
			version text,
			module text,
			arch text,
			os text,
			start text,
			end text );`,

		`DROP TABLE IF EXISTS %v;
		 CREATE TABLE %v (
			id integer not null primary key,
			date text, severity text,
			component text,
			context text,
			msg text,
			plan text,
			type text,
			ns text,
			message text collate nocase,
			op text,
			filter text,
			_index text,
			milli integer,
			reslen integer );`,

		`DROP TABLE IF EXISTS %v_audit;
		 CREATE TABLE %v_audit (
			type text,
			name text,
			value integer );`,

		`DROP TABLE IF EXISTS %v_clients;
		 CREATE TABLE %v_clients (
			id integer not null primary key,
			ip text,
			port text,
			conns integer,
			accepted integer,
			ended integer,
			context text );`,

		`DROP TABLE IF EXISTS %v_drivers;
		 CREATE TABLE %v_drivers (
			id integer not null primary key, 
			ip text, 
			driver text, 
			version text );`,

		`DROP TABLE IF EXISTS %v_ops;
		 CREATE TABLE %v_ops (
			op text,
			count integer,
			avg_ms numeric,
			max_ms integer,
			total_ms integer,
			ns text,
			_index text,
			reslen integer,
			filter text );`,
	}
	for i, table := range tables {
		stmt := table
		if i > 0 {
			stmt = fmt.Sprintf(table, hatchetName, hatchetName)
		}
		log.Println(stmt)
		if _, err = db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// CreateIndexes returns init statement
func CreateIndexes(db *sql.DB, hatchetName string) error {
	var err error
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS %v_idx_component_severity ON %v (component,severity);",
		"CREATE INDEX IF NOT EXISTS %v_idx_context_date ON %v (context,date);",
		"CREATE INDEX IF NOT EXISTS %v_idx_context_op_reslen ON %v (context,op,reslen);",
		"CREATE INDEX IF NOT EXISTS %v_idx_milli ON %v (milli);",
		"CREATE INDEX IF NOT EXISTS %v_idx_ns_reslen ON %v (ns,reslen);",

		"CREATE INDEX IF NOT EXISTS %v_idx_ns_op_filter ON %v (ns,op,filter);",
		"CREATE INDEX IF NOT EXISTS %v_idx_severity ON %v (severity);",

		"CREATE INDEX IF NOT EXISTS %v_audit_idx_type_value ON %v_audit (type,value DESC);",
		"CREATE INDEX IF NOT EXISTS %v_clients_idx_ip_accepted ON %v_clients (ip,accepted);",
		"CREATE INDEX IF NOT EXISTS %v_clients_idx_ip_context ON %v_clients (ip,context);",
		"CREATE INDEX IF NOT EXISTS %v_drivers_idx_driver_version_ip ON %v_drivers (driver,version DESC,ip);",
		"CREATE INDEX IF NOT EXISTS %v_ops_idx_avgms ON %v_ops (avg_ms);",

		"CREATE INDEX IF NOT EXISTS %v_ops_idx_index ON %v_ops (_index);",
	}
	for i, index := range indexes {
		stmts := fmt.Sprintf(index, hatchetName, hatchetName)
		log.Printf("%d/%d: %s\n", i+1, len(indexes), stmts)
		if _, err = db.Exec(stmts); err != nil {
			return err
		}
	}
	return nil
}

// GetHatchetPreparedStmt returns prepared statement of the hatchet table
func GetHatchetPreparedStmt(hatchetName string) string {
	return fmt.Sprintf(`INSERT INTO %v (id, date, severity, component, context,
		msg, plan, type, ns, message, op, filter, _index, milli, reslen)
		VALUES(?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?)`, hatchetName)
}

// GetClientPreparedStmt returns prepared statement of clients table
func GetClientPreparedStmt(hatchetName string) string {
	return fmt.Sprintf(`INSERT INTO %v_clients (id, ip, port, conns, accepted, ended, context)
		VALUES(?,?,?,?,?, ?,?)`, hatchetName)
}

// GetDriverPreparedStmt returns prepared statement of drivers table
func GetDriverPreparedStmt(hatchetName string) string {
	return fmt.Sprintf(`INSERT INTO %v_drivers (id, ip, driver, version)
		VALUES(?,?,?,?)`, hatchetName)
}

func explain(db *sql.DB, stmt string) error {
	result, err := db.Query("EXPLAIN QUERY PLAN " + stmt)
	if err != nil {
		return err
	}
	var line, a, b, c string
	for result.Next() {
		result.Scan(&a, &b, &c, &line)
		log.Println("->", line)
	}
	return nil
}
