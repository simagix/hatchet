// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"database/sql"
	"fmt"
)

func getSlowOps(tableName string, orderBy string, order string) ([]OpStat, error) {
	ops := []OpStat{}
	db, err := sql.Open("sqlite3", SQLITE_FILE)
	if err != nil {
		return ops, err
	}
	defer db.Close()
	query := getSlowOpsQuery(tableName, orderBy, order)
	rows, err := db.Query(query)
	if err != nil {
		return ops, err
	}
	defer rows.Close()
	for rows.Next() {
		var op OpStat
		if err = rows.Scan(&op.Command, &op.Count, &op.AvgMilli, &op.MaxMilli, &op.TotalMilli,
			&op.Namespace, &op.Index, &op.Reslen, &op.Filter, &op.Plan); err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	return ops, err
}

func getSlowOpsQuery(tableName string, orderBy string, order string) string {
	query := fmt.Sprintf(`SELECT op, COUNT(*) "count", ROUND(AVG(milli),1) avg_ms, MAX(milli) max_ms, SUM(milli) total_ms,
			ns, _index "index", SUM(reslen) "reslen", filter "query pattern", plan
			FROM %v WHERE op != "" GROUP BY op, ns, filter ORDER BY %v %v`, tableName, orderBy, order)
	return query
}

func getSlowestLogs(tableName string, topN int) ([]string, error) {
	logstrs := []string{}
	query := fmt.Sprintf(`SELECT date, severity, component, context, message
			FROM %v WHERE op != "" ORDER BY milli DESC LIMIT %v`, tableName, topN)
	db, err := sql.Open("sqlite3", SQLITE_FILE)
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
