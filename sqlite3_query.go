// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"fmt"
	"log"
	"strings"
)

const (
	Q_LIMIT = " LIMIT 100"
)

func (ptr *SQLite3DB) GetSlowOps(tableName string, orderBy string, order string, collscan bool) ([]OpStat, error) {
	ops := []OpStat{}
	db := ptr.db
	query := ptr.GetSlowOpsQuery(tableName, orderBy, order, collscan)
	if ptr.verbose {
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

func (ptr *SQLite3DB) GetSlowOpsQuery(tableName string, orderBy string, order string, collscan bool) string {
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

func (ptr *SQLite3DB) GetLogs(tableName string, opts ...string) ([]LegacyLog, error) {
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
				severities := []string{}
				for _, v := range SEVERITIES {
					severities = append(severities, fmt.Sprintf("'%v'", v))
					if v == toks[1] {
						break
					}
				}
				wheres = append(wheres, " severity IN ("+strings.Join(severities, ",")+")")
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
	db := ptr.db
	if ptr.verbose {
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
		return ptr.GetLogsFromMessage(tableName, opts...)
	}
	return docs, err
}

func (ptr *SQLite3DB) GetLogsFromMessage(tableName string, opts ...string) ([]LegacyLog, error) {
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
			wheres = append(wheres, " severity IN ("+strings.Join(sevs, ",")+")")
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
	if ptr.verbose {
		log.Println(query)
	}
	db := ptr.db
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

func (ptr *SQLite3DB) GetSlowestLogs(tableName string, topN int) ([]string, error) {
	logstrs := []string{}
	query := fmt.Sprintf(`SELECT date, severity, component, context, message
			FROM %v WHERE op != "" ORDER BY milli DESC LIMIT %v`, tableName, topN)
	db := ptr.db
	if ptr.verbose {
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

func (ptr *SQLite3DB) GetSubStringFromTable(tableName string) string {
	substr := "SUBSTR(date, 1, 16)"
	query := fmt.Sprintf(`SELECT start, end FROM hatchet WHERE name = '%v'`, tableName)
	db := ptr.db
	if ptr.verbose {
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
		return GetDateSubString(start, end)
	}
	return substr
}

func (ptr *SQLite3DB) GetAverageOpTime(tableName string, duration string) ([]OpCount, error) {
	docs := []OpCount{}
	db := ptr.db
	substr := ptr.GetSubStringFromTable(tableName)
	durcond := ""
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = GetDateSubString(toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT %v, AVG(milli), COUNT(*), op, ns, filter FROM %v 
		WHERE op != '' %v GROUP by %v, op, ns, filter;`, substr, tableName, durcond, substr)
	if ptr.verbose {
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

func (ptr *SQLite3DB) GetSlowOpsCounts(tableName string, duration string) ([]OpCount, error) {
	docs := []OpCount{}
	db := ptr.db
	substr := ptr.GetSubStringFromTable(tableName)
	durcond := ""
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = GetDateSubString(toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT %v, COUNT(op), op, ns, filter FROM %v 
		WHERE op != '' %v GROUP by %v, op, ns, filter;`, substr, tableName, durcond, substr)
	if ptr.verbose {
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

func (ptr *SQLite3DB) GetTableSummary(tableName string) (string, string, string) {
	query := fmt.Sprintf("SELECT name, version, module, os, arch, start, end FROM hatchet WHERE name = '%v'", tableName)
	db := ptr.db
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

func (ptr *SQLite3DB) GetTables() ([]string, error) {
	tables := []string{}
	query := "SELECT name, version, module, os, arch FROM hatchet ORDER BY name"
	db := ptr.db
	if ptr.verbose {
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

// GetAcceptedConnsCounts returns opened connection counts
func (ptr *SQLite3DB) GetAcceptedConnsCounts(tableName string, duration string) ([]NameValue, error) {
	docs := []NameValue{}
	var durcond string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT b.ip, SUM(b.accepted)
		FROM %v a, %v_rmt b WHERE a.id = b.id AND b.accepted = 1 %v GROUP by ip ORDER BY accepted DESC;`, tableName, tableName, durcond)
	db := ptr.db
	if ptr.verbose {
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

// GetConnectionStats returns stats data of accepted and ended
func (ptr *SQLite3DB) GetConnectionStats(tableName string, chartType string, duration string) ([]Remote, error) {
	docs := []Remote{}
	var query, durcond string
	substr := ptr.GetSubStringFromTable(tableName)
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = GetDateSubString(toks[0], toks[1])
	}
	if chartType == "time" {
		query = fmt.Sprintf(`SELECT %v dt, SUM(b.accepted), SUM(b.ended)
			FROM %v a, %v_rmt b WHERE a.id = b.id %v GROUP by dt ORDER BY dt;`,
			substr, tableName, tableName, durcond)
	} else if chartType == "total" {
		query = fmt.Sprintf(`SELECT b.ip, SUM(b.accepted), SUM(b.ended)
			FROM %v a, %v_rmt b WHERE a.id = b.id %v GROUP by ip ORDER BY accepted DESC;`, tableName, tableName, durcond)
	}
	db := ptr.db
	if ptr.verbose {
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
		if err = rows.Scan(&doc.Value, &accepted, &ended); err != nil {
			return docs, err
		}
		doc.Accepted = int(accepted)
		doc.Ended = int(ended)
		docs = append(docs, doc)
	}
	return docs, err
}

// GetOpsCounts returns opened connection counts
func (ptr *SQLite3DB) GetOpsCounts(tableName string, duration string) ([]NameValue, error) {
	docs := []NameValue{}
	var durcond string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT op, COUNT(op) counts
		FROM %v WHERE op != '' %v GROUP by op ORDER BY counts DESC;`, tableName, durcond)
	db := ptr.db
	if ptr.verbose {
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
