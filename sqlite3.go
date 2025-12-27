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
	"strings"
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
	// Run schema migrations
	if err = sqlite.migrate(); err != nil {
		log.Println("warning: migration failed:", err)
	}
	return sqlite, nil
}

// migrate checks and applies schema migrations for backward compatibility
func (ptr *SQLite3DB) migrate() error {
	// Check if hatchet table exists
	var count int
	err := ptr.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='hatchet'").Scan(&count)
	if err != nil || count == 0 {
		return nil // Table doesn't exist yet, will be created later
	}

	// Check if created_at column exists
	rows, err := ptr.db.Query("PRAGMA table_info(hatchet)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasCreatedAt := false
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "created_at" {
			hasCreatedAt = true
			break
		}
	}

	// Add created_at column if missing
	if !hasCreatedAt {
		log.Println("migrating: adding created_at column to hatchet table")
		_, err = ptr.db.Exec("ALTER TABLE hatchet ADD COLUMN created_at text")
		if err != nil {
			return err
		}
	}
	return nil
}

func (ptr *SQLite3DB) GetVerbose() bool {
	return ptr.verbose
}

func (ptr *SQLite3DB) SetVerbose(b bool) {
	ptr.verbose = b
}

func (ptr *SQLite3DB) Begin() error {
	log.Println("creating hatchet", ptr.hatchetName)
	// Drop existing tables first to allow overwriting
	if err := ptr.Drop(); err != nil {
		log.Println("warning: failed to drop existing tables:", err)
	}
	stmts, err := CreateTables(ptr.db, ptr.hatchetName)
	if err != nil {
		return err
	}
	if ptr.verbose {
		for _, stmt := range stmts {
			log.Println(stmt)
		}
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
			DROP INDEX IF EXISTS %v_idx_date_op_ns;
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
		hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName, hatchetName,
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
		doc.Msg, doc.Attributes.PlanSummary, BsonD2M(doc.Attr)["type"], doc.Attributes.NS, doc.Message,
		stat.Op, stat.QueryPattern, stat.Index, doc.Attributes.Milli, doc.Attributes.Reslen,
		doc.Attributes.AppName, doc.Marker)
	return err
}

func (ptr *SQLite3DB) InsertClientConn(index int, doc *Logv2Info) error {
	var err error
	client := doc.Client
	_, err = ptr.clientStmt.Exec(index, client.IP, client.Port, client.Conns, client.Accepted,
		client.Ended, doc.Context, doc.Marker)
	return err
}

func (ptr *SQLite3DB) InsertDriver(index int, doc *Logv2Info) error {
	var err error
	client := doc.Client
	_, err = ptr.driverStmt.Exec(index, client.IP, client.Driver, client.Version, doc.Marker)
	return err
}

func (ptr *SQLite3DB) InsertFailedMessages(m *FailedMessages) error {
	var err error
	for k, v := range m.counters {
		str := strings.ReplaceAll(k, "'", "''")
		stmt := fmt.Sprintf("INSERT INTO %v_audit (type, name, value) VALUES ('failed','%s', %d)", ptr.hatchetName, str, v)
		if _, err = ptr.db.Exec(stmt); err != nil {
			log.Println("error", err, "stmt", stmt, "(k,v)", k, v)
			return err
		}
	}
	return nil
}

func (ptr *SQLite3DB) UpdateHatchetInfo(info HatchetInfo) error {
	istmt := fmt.Sprintf(`INSERT OR REPLACE INTO hatchet (name, version, module, arch, os, start, end, merge, created_at)
		VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v', %v, datetime('now'));`,
		ptr.hatchetName, info.Version, info.Module, info.Arch, info.OS, info.Start, info.End, info.Merge)
	_, err := ptr.db.Exec(istmt)
	return err
}

func (ptr *SQLite3DB) CreateMetaData() error {
	log.Println("creating indexes and this may take minutes")
	stmts, err := CreateIndexes(ptr.db, ptr.hatchetName)
	if err != nil {
		return err
	}
	if ptr.verbose {
		for i, stmt := range stmts {
			log.Printf("%d/%d: %s\n", i+1, len(stmts), stmt)
		}
	}
	// create an index to speed up the Average Operation Time chart
	info := ptr.GetHatchetInfo()
	substr := GetSQLDateSubString(info.Start, info.End)
	toks := strings.Split(substr, "||")
	groupby := substr
	if len(toks) > 1 {
		groupby = toks[0]
	}
	stmt := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %v_idx_date_op_ns ON %v (%v,op,ns,filter,milli);", ptr.hatchetName, ptr.hatchetName, groupby)
	if ptr.verbose {
		log.Printf("%s\n", stmt)
	}
	if _, err = ptr.db.Exec(stmt); err != nil {
		return err
	}

	log.Printf("insert ops into %v_ops\n", ptr.hatchetName)
	query := fmt.Sprintf(`INSERT INTO %v_ops
			SELECT op, COUNT(*), ROUND(AVG(milli),1), MAX(milli), SUM(milli), ns, _index, SUM(reslen), filter, marker
				FROM %v WHERE op != "" GROUP BY op, ns, filter, _index, marker`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [exception] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'exception', severity, COUNT(*) count FROM %v WHERE severity IN ('W', 'E', 'F')
		GROUP by severity`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [op] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'op', op, COUNT(*) count FROM %v WHERE op != '' GROUP by op`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [ip] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'ip', ip, SUM(accepted) open FROM %v_clients GROUP by ip`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [ns] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'ns', ns, COUNT(*) count FROM %v WHERE op != "" GROUP by ns`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [reslen-ns] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'reslen-ns', ns, SUM(reslen) FROM %v WHERE ns != "" AND reslen > 0 GROUP by ns`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [reslen-ip] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'reslen-ip', ip, SUM(reslen) FROM (
			SELECT a.context, sum(reslen) reslen, b.ip ip FROM %v a, %v_clients b
				WHERE op != "" and reslen > 0 and a.context = b.context GROUP by a.context
		) GROUP BY ip`,
		ptr.hatchetName, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [ended-ip] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'ended-ip', ip, SUM(ended) FROM %v_clients
		GROUP BY ip`,
		ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [appname] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'appname', appname, COUNT(*) count FROM %v WHERE appname != "" GROUP by appname`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}

	log.Printf("insert [reslen-appname] into %v_audit\n", ptr.hatchetName)
	query = fmt.Sprintf(`INSERT INTO %v_audit
		SELECT 'reslen-appname', appname, SUM(reslen) FROM %v WHERE appname != "" AND reslen > 0 GROUP by appname`, ptr.hatchetName, ptr.hatchetName)
	if ptr.verbose {
		explain(ptr.db, query)
	}
	if _, err = ptr.db.Exec(query); err != nil {
		return err
	}
	return err
}

// CreateTables returns init statement
func CreateTables(db *sql.DB, hatchetName string) ([]string, error) {
	var err error
	tables := []string{
		`CREATE TABLE IF NOT EXISTS hatchet (
			name text not null primary key,
			version text,
			module text,
			arch text,
			os text,
			start text,
			end text,
			merge integer,
			created_at text);`,

		`CREATE TABLE IF NOT EXISTS %v (
			id integer not null,
			date text,
			severity text,
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
			reslen integer,
			appname text,
			marker integer);`,

		`CREATE TABLE IF NOT EXISTS %v_audit (
			type text,
			name text,
			value integer );`,

		`CREATE TABLE IF NOT EXISTS %v_clients (
			id integer not null,
			ip text,
			port text,
			conns integer,
			accepted integer,
			ended integer,
			context text,
			marker integer);`,

		`CREATE TABLE IF NOT EXISTS %v_drivers (
			id integer not null,
			ip text,
			driver text,
			version text,
			marker integer);`,

		`CREATE TABLE IF NOT EXISTS %v_ops (
			op text,
			count integer,
			avg_ms numeric,
			max_ms integer,
			total_ms integer,
			ns text,
			_index text,
			reslen integer,
			filter text,
			marker integer);`,
	}
	stmts := []string{}
	for i, table := range tables {
		stmt := table
		if i > 0 {
			stmt = fmt.Sprintf(table, hatchetName)
		}
		stmts = append(stmts, stmt)
		if _, err = db.Exec(stmt); err != nil {
			return nil, err
		}
	}
	return stmts, nil
}

// CreateIndexes returns init statement
func CreateIndexes(db *sql.DB, hatchetName string) ([]string, error) {
	var err error
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS %v_idx_component_severity ON %v (component,severity);",
		"CREATE INDEX IF NOT EXISTS %v_idx_context_date ON %v (context,date);",
		"CREATE INDEX IF NOT EXISTS %v_idx_context_op_reslen ON %v (context,op,reslen);",
		"CREATE INDEX IF NOT EXISTS %v_idx_milli ON %v (milli);",
		"CREATE INDEX IF NOT EXISTS %v_idx_ns_reslen ON %v (ns,reslen);",

		"CREATE INDEX IF NOT EXISTS %v_idx_ns_op_filter ON %v (ns,op,filter);",
		"CREATE INDEX IF NOT EXISTS %v_idx_severity ON %v (severity);",
		"CREATE INDEX IF NOT EXISTS %v_idx_appname_reslen ON %v (appname,reslen);",

		"CREATE INDEX IF NOT EXISTS %v_audit_idx_type_value ON %v_audit (type,value DESC);",
		"CREATE INDEX IF NOT EXISTS %v_clients_idx_ip_accepted ON %v_clients (ip,accepted);",
		"CREATE INDEX IF NOT EXISTS %v_clients_idx_ip_context ON %v_clients (ip,context);",
		"CREATE INDEX IF NOT EXISTS %v_drivers_idx_driver_version_ip ON %v_drivers (driver,version DESC,ip);",
		"CREATE INDEX IF NOT EXISTS %v_ops_idx_avgms ON %v_ops (avg_ms);",

		"CREATE INDEX IF NOT EXISTS %v_ops_idx_index ON %v_ops (_index);",
	}
	stmts := []string{}
	for _, index := range indexes {
		stmt := fmt.Sprintf(index, hatchetName, hatchetName)
		stmts = append(stmts, stmt)
		if _, err = db.Exec(stmt); err != nil {
			return nil, err
		}
	}
	return stmts, nil
}

// GetHatchetPreparedStmt returns prepared statement of the hatchet table
func GetHatchetPreparedStmt(hatchetName string) string {
	return fmt.Sprintf(`INSERT INTO %v (id, date, severity, component, context,
		msg, plan, type, ns, message, op, filter, _index, milli, reslen, appname, marker)
		VALUES(?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?, ?,?)`, hatchetName)
}

// GetClientPreparedStmt returns prepared statement of clients table
func GetClientPreparedStmt(hatchetName string) string {
	return fmt.Sprintf(`INSERT INTO %v_clients (id, ip, port, conns, accepted, ended, context, marker)
		VALUES(?,?,?,?,?, ?,?,?)`, hatchetName)
}

// GetDriverPreparedStmt returns prepared statement of drivers table
func GetDriverPreparedStmt(hatchetName string) string {
	return fmt.Sprintf(`INSERT INTO %v_drivers (id, ip, driver, version, marker)
		VALUES(?,?,?,?,?)`, hatchetName)
}

func explain(db *sql.DB, query string) error {
	explainIt := "EXPLAIN QUERY PLAN " + query
	log.Println(explainIt)
	result, err := db.Query(explainIt)
	if err != nil {
		return err
	}
	defer result.Close()
	var line, a, b, c string
	for result.Next() {
		result.Scan(&a, &b, &c, &line)
		log.Println("->", line)
	}
	return nil
}
