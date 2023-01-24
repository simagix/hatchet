// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	Q_LIMIT = " LIMIT 100"
)

// getHatchetInitStmt returns init statement
func getHatchetInitStmt(tableName string) string {
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
	if GetLogv2().verbose {
		log.Println(query)
	}
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
	query := fmt.Sprintf(`SELECT op, count, avg_ms, max_ms,
			total_ms, ns, _index "index", reslen, filter "query pattern"
			FROM %v_ops ORDER BY %v %v`, tableName, orderBy, order)
	if collscan {
		query = fmt.Sprintf(`SELECT op, count, avg_ms, max_ms,
				total_ms, ns, _index "index", reslen, filter "query pattern"
				FROM %v_ops WHERE _index = "COLLSCAN" ORDER BY %v %v`, tableName, orderBy, order)
	}
	return query
}

func getLogs(tableName string, opts ...string) ([]LegacyLog, error) {
	docs := []LegacyLog{}
	qheader := fmt.Sprintf(`SELECT date, severity, component, context, message FROM %v`, tableName)
	wheres := []string{}
	search := ""
	qlimit := Q_LIMIT

	if len(opts) > 0 {
		for _, opt := range opts {
			toks := strings.Split(opt, "=")
			if len(toks) < 2 || toks[1] == "" {
				continue
			}
			if toks[0] == "duration" {
				dates := strings.Split(toks[1], ",")
				wheres = append(wheres, fmt.Sprintf(" date BETWEEN '%v' and '%v'", dates[0], dates[1]))
			} else if toks[0] == "limit" {
				qlimit = " LIMIT " + toks[1]
			} else if toks[0] == "severity" {
				sevs := []string{}
				for _, v := range SEVERITIES {
					sevs = append(sevs, fmt.Sprintf("'%v'", v))
					if v == toks[1] {
						break
					}
				}
				wheres = append(wheres, " severity IN (" + strings.Join(sevs, ",") + ")")
			} else {
				wheres = append(wheres, fmt.Sprintf(` %v = "%v"`, toks[0], EscapeString(toks[1])))
				if toks[0] == "context" {
					search = toks[1]
				}
			}
		}
	}
	wclause := ""
	if len(wheres) > 0 {
		wclause = " WHERE " + strings.Join(wheres, " AND")
	}
	query := qheader + wclause + qlimit
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	if GetLogv2().verbose {
		log.Println(query)
	}
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
	if len(docs) == 0 && search != "" { // no context found, perform message search
		return getLogsFromMessage(tableName, opts...)
	}
	return docs, err
}

func getLogsFromMessage(tableName string, opts ...string) ([]LegacyLog, error) {
	qheader := fmt.Sprintf(`SELECT date, severity, component, context, message FROM %v`, tableName)
	docs := []LegacyLog{}
	wheres := []string{}
	qlimit := Q_LIMIT
	for _, opt := range opts {
		toks := strings.Split(opt, "=")
		if len(toks) < 2 || toks[1] == "" {
			continue
		}
		if toks[0] == "duration" {
			dates := strings.Split(toks[1], ",")
			wheres = append(wheres, fmt.Sprintf(" date BETWEEN '%v' and '%v'", dates[0], dates[1]))
		} else if toks[0] == "limit" {
			qlimit = " LIMIT " + toks[1]
		} else if toks[0] == "severity" {
			sevs := []string{}
			for _, v := range SEVERITIES {
				sevs = append(sevs, fmt.Sprintf("'%v'", v))
				if v == toks[1] {
					break
				}
			}
			wheres = append(wheres, " severity IN (" + strings.Join(sevs, ",") + ")")
		} else if toks[0] == "context" {
			wheres = append(wheres, fmt.Sprintf(` LOWER(message) LIKE "%%%v%%"`, EscapeString(toks[1])))
		} else {
			wheres = append(wheres, fmt.Sprintf(` %v = "%v"`, toks[0], EscapeString(toks[1])))
		}
	}
	wclause := ""
	if len(wheres) > 0 {
		wclause = " WHERE " + strings.Join(wheres, " AND")
	}
	query := qheader + wclause + qlimit
	if GetLogv2().verbose {
		log.Println(query)
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
	if GetLogv2().verbose {
		log.Println(query)
	}
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
	Milli     float64
	Op        string
	Namespace string
	Filter    string
}

func getSubStringFromTable(tableName string) string {
	substr := "SUBSTR(date, 1, 16)"
	query := fmt.Sprintf(`SELECT start, end FROM hatchet WHERE name = '%v'`, tableName)
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return substr
	}
	defer db.Close()
	if GetLogv2().verbose {
		log.Println(query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return substr
	}
	defer rows.Close()
	if rows.Next() {
		var start, end string
		if err = rows.Scan(&start, &end); err != nil {
			return substr
		}
		return getDateSubString(start, end)
	}
	return substr
}

func getDateSubString(start string, end string) string {
	var err error
	substr := "SUBSTR(date, 1, 16)"
	var stime, etime time.Time
	if stime, err = time.Parse(time.RFC3339, start); err != nil {
		return substr
	}
	if etime, err = time.Parse(time.RFC3339, end); err != nil {
		return substr
	}
	minutes := etime.Sub(stime).Minutes()
	if minutes < 1 {
		return "SUBSTR(date, 1, 19)"
	} else if minutes < 10 {
		return "SUBSTR(date, 1, 18)||'9'"
	} else if minutes < 60 {
		return "SUBSTR(date, 1, 16)||':59'"
	} else if minutes < 600 {
		return "SUBSTR(date, 1, 15)||'9:59'"
	} else if minutes < 3600 {
		return "SUBSTR(date, 1, 13)||':59:59'"
	} else {
		return "SUBSTR(date, 1, 10)||'T23:59:59'"
	}
}

func getAverageOpTime(tableName string, duration string) ([]OpCount, error) {
	docs := []OpCount{}
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	substr := getSubStringFromTable(tableName)
	durcond := ""
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = getDateSubString(toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT %v, AVG(milli), COUNT(*), op, ns, filter FROM %v 
		WHERE op != '' %v GROUP by %v, op, ns, filter;`, substr, tableName, durcond, substr)
	if GetLogv2().verbose {
		log.Println(query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return docs, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc OpCount
		if err = rows.Scan(&doc.Date, &doc.Milli, &doc.Count, &doc.Op, &doc.Namespace, &doc.Filter); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	return docs, err
}

func getSlowOpsCounts(tableName string, duration string) ([]OpCount, error) {
	docs := []OpCount{}
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	substr := getSubStringFromTable(tableName)
	durcond := ""
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = getDateSubString(toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT %v, COUNT(op), op, ns, filter FROM %v 
		WHERE op != '' %v GROUP by %v, op, ns, filter;`, substr, tableName, durcond, substr)
	if GetLogv2().verbose {
		log.Println(query)
	}
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

func getTableSummary(tableName string) (string, string, string) {
	query := fmt.Sprintf("SELECT name, version, module, os, arch, start, end FROM hatchet WHERE name = '%v'", tableName)
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return "", "", ""
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return "", "", ""
	}
	defer rows.Close()
	if rows.Next() {
		var table, version, module, os, arch, start, end string
		if err = rows.Scan(&table, &version, &module, &os, &arch, &start, &end); err != nil {
			return "", "", ""
		}
		arr := []string{}
		if module == "" {
			module = "community"
		}
		if version != "" {
			arr = append(arr, fmt.Sprintf(": MongoDB v%v (%v)", version, module))
		}
		if os != "" {
			arr = append(arr, "os: "+os)
		}
		if arch != "" {
			arr = append(arr, "arch: "+arch)
		}
		return table + strings.Join(arr, ", "), start, end
	}
	return "", "", ""
}

func getTables() ([]string, error) {
	tables := []string{}
	query := "SELECT name, version, module, os, arch FROM hatchet ORDER BY name"
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return tables, err
	}
	defer db.Close()
	if GetLogv2().verbose {
		log.Println(query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return tables, err
	}
	defer rows.Close()
	for rows.Next() {
		var table, version, module, os, arch string
		if err = rows.Scan(&table, &version, &module, &os, &arch); err != nil {
			return tables, err
		}
		tables = append(tables, table)
	}
	return tables, err
}

// getAcceptedConnsCounts returns opened connection counts
func getAcceptedConnsCounts(tableName string, duration string) ([]NameValue, error) {
	docs := []NameValue{}
	var durcond string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT b.ip, SUM(b.accepted)
		FROM %v a, %v_rmt b WHERE a.id = b.id AND b.accepted = 1 %v GROUP by ip ORDER BY accepted DESC;`, tableName, tableName, durcond)
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	if GetLogv2().verbose {
		log.Println(query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return docs, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc NameValue
		var conns float64
		if err = rows.Scan(&doc.Name, &conns); err != nil {
			return docs, err
		}
		doc.Value = int(conns)
		docs = append(docs, doc)
	}
	return docs, err
}

// getConnectionStats returns stats data of accepted and ended
func getConnectionStats(tableName string, chartType string, duration string) ([]Remote, error) {
	docs := []Remote{}
	var query, durcond string
	substr := getSubStringFromTable(tableName)
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = getDateSubString(toks[0], toks[1])
	}
	if chartType == "time" {
		query = fmt.Sprintf(`SELECT %v dt, SUM(b.accepted), SUM(b.ended)
			FROM %v a, %v_rmt b WHERE a.id = b.id %v GROUP by dt ORDER BY dt;`,
			substr, tableName, tableName, durcond)
	} else if chartType == "total" {
		query = fmt.Sprintf(`SELECT b.ip, SUM(b.accepted), SUM(b.ended)
			FROM %v a, %v_rmt b WHERE a.id = b.id %v GROUP by ip ORDER BY accepted DESC;`, tableName, tableName, durcond)
	}
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	if GetLogv2().verbose {
		log.Println(query)
	}
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

// getOpsCounts returns opened connection counts
func getOpsCounts(tableName string, duration string) ([]NameValue, error) {
	docs := []NameValue{}
	var durcond string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT op, COUNT(op) counts
		FROM %v WHERE op != '' %v GROUP by op ORDER BY counts DESC;`, tableName, durcond)
	db, err := sql.Open("sqlite3", GetLogv2().dbfile)
	if err != nil {
		return docs, err
	}
	defer db.Close()
	if GetLogv2().verbose {
		log.Println(query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return docs, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc NameValue
		var conns float64
		if err = rows.Scan(&doc.Name, &conns); err != nil {
			return docs, err
		}
		doc.Value = int(conns)
		docs = append(docs, doc)
	}
	return docs, err
}
