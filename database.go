// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"database/sql"
	"fmt"
	"strings"
)

// getHatchetInitStmt returns init statement
func getHatchetInitStmt(tableName string) string {
	return fmt.Sprintf(`
			DROP TABLE IF EXISTS %v;
			CREATE TABLE %v (
				id integer not null primary key, date text, severity text, component text, context text,
				msg text, plan text, type text, ns text, message text,
				op text, filter text, _index text, milli integer, reslen integer);
			CREATE INDEX IF NOT EXISTS %v_idx_component ON %v (component);
			CREATE INDEX IF NOT EXISTS %v_idx_context ON %v (context);
			CREATE INDEX IF NOT EXISTS %v_idx_severity ON %v (severity);
			CREATE INDEX IF NOT EXISTS %v_idx_op ON %v (op,ns,filter);

			DROP TABLE IF EXISTS %v_rmt;
			CREATE TABLE %v_rmt(
				id integer not null primary key, ip text, port text, conns integer, accepted integer, ended integer)`,
		tableName, tableName, tableName, tableName, tableName,
		tableName, tableName, tableName, tableName, tableName,
		tableName, tableName)
}

// getHatchetPreparedStmt returns prepared statement of log table
func getHatchetPreparedStmt(id string) string {
	return fmt.Sprintf(`INSERT INTO %v (id, date, severity, component, context,
		msg, plan, type, ns, message, op, filter, _index, milli, reslen)
		VALUES(?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?)`, id)
}

// getClientPreparedStmt returns prepared statement of client table
func getClientPreparedStmt(id string) string {
	return fmt.Sprintf(`INSERT INTO %v_rmt (id, ip, port, conns, accepted, ended)
		VALUES(?,?,?,?,?, ?)`, id)
}

func getSlowOps(tableName string, orderBy string, order string, collscan bool) ([]OpStat, error) {
	ops := []OpStat{}
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return ops, err
	}
	defer db.Close()
	query := getSlowOpsQuery(tableName, orderBy, order, collscan)
	rows, err := db.Query(query)
	if err != nil {
		return ops, err
	}
	defer rows.Close()
	for rows.Next() {
		var op OpStat
		if err = rows.Scan(&op.Op, &op.Count, &op.AvgMilli, &op.MaxMilli, &op.TotalMilli,
			&op.Namespace, &op.Index, &op.Reslen, &op.QueryPattern); err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	return ops, err
}

func getSlowOpsQuery(tableName string, orderBy string, order string, collscan bool) string {
	query := fmt.Sprintf(`SELECT op, COUNT(*) "count", ROUND(AVG(milli),1) avg_ms, MAX(milli) max_ms,
			SUM(milli) total_ms, ns, _index "index", SUM(reslen) "reslen", filter "query pattern"
			FROM %v WHERE op != "" GROUP BY op, ns, filter ORDER BY %v %v`, tableName, orderBy, order)
	if collscan {
		query = fmt.Sprintf(`SELECT op, COUNT(*) "count", ROUND(AVG(milli),1) avg_ms, MAX(milli) max_ms,
				SUM(milli) total_ms, ns, _index "index", SUM(reslen) "reslen", filter "query pattern"
				FROM %v WHERE op != "" AND _index = "COLLSCAN" GROUP BY op, ns, filter ORDER BY %v %v`,
			tableName, orderBy, order)
	}
	return query
}

func getLogs(tableName string, opts ...string) ([]LegacyLog, error) {
	docs := []LegacyLog{}
	query := fmt.Sprintf(`SELECT date, severity, component, context, message FROM %v`, tableName)
	if len(opts) > 0 {
		cnt := 0
		for _, opt := range opts {
			toks := strings.Split(opt, "=")
			if len(toks) < 2 || toks[1] == "" {
				continue
			}
			if cnt == 0 {
				query += " WHERE"
			} else if cnt > 0 {
				query += " AND"
			}
			if toks[0] == "duration" {
				dates := strings.Split(toks[1], ",")
				query += fmt.Sprintf(" date BETWEEN '%v' and '%v'", dates[0], dates[1])
			} else {
				query += fmt.Sprintf(" %v = '%v'", toks[0], toks[1])
			}
			cnt++
		}
	}
	query += " LIMIT 1000"
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return docs, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc LegacyLog
		if err = rows.Scan(&doc.Timestamp, &doc.Severity, &doc.Component, &doc.Context, &doc.Message); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	return docs, err
}

func getSlowestLogs(tableName string, topN int) ([]string, error) {
	logstrs := []string{}
	query := fmt.Sprintf(`SELECT date, severity, component, context, message
			FROM %v WHERE op != "" ORDER BY milli DESC LIMIT %v`, tableName, topN)
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return logstrs, err
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return logstrs, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc Logv2Info
		var date string
		if err = rows.Scan(&date, &doc.Severity, &doc.Component, &doc.Context, &doc.Message); err != nil {
			return logstrs, err
		}
		logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", date, doc.Severity, doc.Component, doc.Context, doc.Message)
		logstrs = append(logstrs, logstr)
	}
	return logstrs, err
}

type OpCount struct {
	Date      string
	Count     int
	Op        string
	Namespace string
	Filter    string
}

func getOpCounts(tableName string) ([]OpCount, error) {
	docs := []OpCount{}
	query := fmt.Sprintf(`SELECT SUBSTR(date, 1, 16), COUNT(op), op, ns, filter 
		FROM %v WHERE op != ''
		GROUP by SUBSTR(date, 1, 16), op, ns, filter;`, tableName)
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return docs, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc OpCount
		if err = rows.Scan(&doc.Date, &doc.Count, &doc.Op, &doc.Namespace, &doc.Filter); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	return docs, err
}

func getTables() ([]string, error) {
	tables := []string{}
	query := "SELECT tbl_name FROM sqlite_schema WHERE type = 'table' ORDER BY tbl_name"
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return tables, err
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return tables, err
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			return tables, err
		}
		if strings.HasSuffix(table, "_rmt") {
			continue
		}
		tables = append(tables, table)
	}
	return tables, err
}

// getAcceptedConnsCount returns opened connection counts
func getAcceptedConnsCount(tableName string) ([]Remote, error) {
	docs := []Remote{}
	query := fmt.Sprintf(`SELECT ip, SUM(accepted)
		FROM %v_rmt WHERE accepted = 1 GROUP by ip ORDER BY accepted DESC;`, tableName)
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return docs, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc Remote
		var conns float64
		if err = rows.Scan(&doc.IP, &conns); err != nil {
			return docs, err
		}
		doc.Conns = int(conns)
		docs = append(docs, doc)
	}
	return docs, err
}

// getConnectionStats returns stats data of accepted and ended
func getConnectionStats(tableName string, chartType string) ([]Remote, error) {
	docs := []Remote{}
	query := fmt.Sprintf(`SELECT ip, SUM(accepted), SUM(ended)
		FROM %v_rmt GROUP by ip ORDER BY accepted DESC;`, tableName)
	if chartType == "time" {
		query = fmt.Sprintf(`SELECT SUBSTR(a.date, 1, 16) dt, SUM(b.accepted), SUM(b.ended)
			FROM %v a, %v_rmt b WHERE a.id = b.id GROUP by dt ORDER BY dt;`,
			tableName, tableName)
	}
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return docs, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc Remote
		var accepted float64
		var ended float64
		if err = rows.Scan(&doc.IP, &accepted, &ended); err != nil {
			return docs, err
		}
		doc.Accepted = int(accepted)
		doc.Ended = int(ended)
		docs = append(docs, doc)
	}
	return docs, err
}
