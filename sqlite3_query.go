/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * sqlite3_query.go
 */

package hatchet

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

type OpCount struct {
	Date      string  `bson:"date"`
	Count     int     `bson:"count"`
	Milli     float64 `bson:"milli"`
	Op        string  `bson:"op"`
	Namespace string  `bson:"ns"`
	Filter    string  `bson:"filter"`
}

func (ptr *SQLite3DB) GetSlowOps(orderBy string, order string, collscan bool) ([]OpStat, error) {
	ops := []OpStat{}
	db := ptr.db
	query := fmt.Sprintf(`SELECT op, count, avg_ms, max_ms,
			total_ms, ns, _index "index", reslen, filter "query_pattern"
			FROM %v_ops ORDER BY %v %v`, ptr.hatchetName, orderBy, order)
	if collscan {
		query = fmt.Sprintf(`SELECT op, count, avg_ms, max_ms,
				total_ms, ns, _index "index", reslen, filter "query_pattern"
				FROM %v_ops WHERE _index = "COLLSCAN" ORDER BY %v %v`, ptr.hatchetName, orderBy, order)
	}
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

func (ptr *SQLite3DB) GetLogs(opts ...string) ([]LegacyLog, error) {
	docs := []LegacyLog{}
	qheader := fmt.Sprintf(`SELECT date, severity, component, context, message FROM %v`, ptr.hatchetName)
	wheres := []string{}
	search := ""
	qlimit := LIMIT + 1
	var offset, nlimit int

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
				offset, nlimit = GetOffsetLimit(toks[1])
				qlimit = ToInt(nlimit) + 1
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
	query := qheader + wclause + fmt.Sprintf(" LIMIT %v,%v", offset, qlimit)
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
		return ptr.SearchLogs(opts...)
	}
	return docs, err
}

func (ptr *SQLite3DB) SearchLogs(opts ...string) ([]LegacyLog, error) {
	qheader := fmt.Sprintf(`SELECT date, severity, component, context, message FROM %v`, ptr.hatchetName)
	docs := []LegacyLog{}
	wheres := []string{}
	qlimit := LIMIT + 1
	var offset, nlimit int
	for _, opt := range opts {
		toks := strings.Split(opt, "=")
		if len(toks) < 2 || toks[1] == "" {
			continue
		}
		if toks[0] == "duration" {
			dates := strings.Split(toks[1], ",")
			wheres = append(wheres, fmt.Sprintf(" date BETWEEN '%v' and '%v'", dates[0], dates[1]))
		} else if toks[0] == "limit" {
			offset, nlimit = GetOffsetLimit(toks[1])
			qlimit = ToInt(nlimit) + 1
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
	query := qheader + wclause + fmt.Sprintf(" LIMIT %v,%v", offset, qlimit)
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

func (ptr *SQLite3DB) GetSlowestLogs(topN int) ([]LegacyLog, error) {
	docs := []LegacyLog{}
	query := fmt.Sprintf(`SELECT date, severity, component, context, message
			FROM %v WHERE op != "" ORDER BY milli DESC LIMIT %v`, ptr.hatchetName, topN)
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
	return docs, err
}

func (ptr *SQLite3DB) GetAverageOpTime(op string, duration string) ([]OpCount, error) {
	docs := []OpCount{}
	db := ptr.db
	durcond := ""
	var substr string
	opcond := "op != ''"
	if op != "" {
		opcond = fmt.Sprintf("op = '%v'", op)
	}
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = GetSQLDateSubString(toks[0], toks[1])
	} else {
		info := ptr.GetHatchetInfo()
		substr = GetSQLDateSubString(info.Start, info.End)
	}
	query := fmt.Sprintf(`SELECT %v, AVG(milli), COUNT(*), op, ns, filter FROM %v 
		WHERE %v %v GROUP by %v, op, ns, filter;`, substr, ptr.hatchetName, opcond, durcond, substr)
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

func (ptr *SQLite3DB) GetHatchetInfo() HatchetInfo {
	var info HatchetInfo
	query := fmt.Sprintf("SELECT name, version, module, os, arch, start, end FROM hatchet WHERE name = '%v'",
		ptr.hatchetName)
	db := ptr.db
	rows, err := db.Query(query)
	if err != nil {
		return info
	}
	if rows.Next() {
		if err = rows.Scan(&info.Name, &info.Version, &info.Module, &info.OS, &info.Arch,
			&info.Start, &info.End); err != nil {
			return info
		}
	}
	if rows != nil {
		rows.Close()
	}

	query = fmt.Sprintf(`SELECT message FROM %v WHERE component = 'CONTROL' AND message LIKE '%%provider:%%region:%%';`,
		ptr.hatchetName)
	if ptr.verbose {
		log.Println(query)
	}
	rows, err = db.Query(query)
	if err == nil && rows.Next() {
		var message string
		if err = rows.Scan(&message); err == nil {
			re := regexp.MustCompile(`.*(provider: "(\w+)", region: "(\w+)",).*`)
			matches := re.FindStringSubmatch(message)
			info.Provider = matches[2]
			info.Region = matches[3]
		}
	}
	if rows != nil {
		rows.Close()
	}

	query = fmt.Sprintf(`SELECT DISTINCT driver, version FROM %v_drivers;`, ptr.hatchetName)
	if ptr.verbose {
		log.Println(query)
	}
	rows, err = db.Query(query)
	for err == nil && rows.Next() {
		var driver, version string
		if err = rows.Scan(&driver, &version); err == nil {
			info.Drivers = append(info.Drivers, map[string]string{driver: version})
		}
	}
	if rows != nil {
		rows.Close()
	}
	return info
}

func (ptr *SQLite3DB) GetHatchetNames() ([]string, error) {
	names := []string{}
	query := "SELECT name FROM hatchet ORDER BY name"
	db := ptr.db
	if ptr.verbose {
		log.Println(query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return names, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return names, err
		}
		names = append(names, name)
	}
	return names, err
}

// GetAcceptedConnsCounts returns opened connection counts
func (ptr *SQLite3DB) GetAcceptedConnsCounts(duration string) ([]NameValue, error) {
	hatchetName := ptr.hatchetName
	docs := []NameValue{}
	var durcond string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT b.ip, SUM(b.accepted)
		FROM %v a, %v_clients b WHERE a.id = b.id AND b.accepted = 1 %v GROUP by ip ORDER BY accepted DESC;`,
		hatchetName, hatchetName, durcond)
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
func (ptr *SQLite3DB) GetConnectionStats(chartType string, duration string) ([]RemoteClient, error) {
	hatchetName := ptr.hatchetName
	docs := []RemoteClient{}
	var query, durcond string
	var substr string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = GetSQLDateSubString(toks[0], toks[1])
	} else {
		info := ptr.GetHatchetInfo()
		substr = GetSQLDateSubString(info.Start, info.End)
	}
	if chartType == "time" {
		query = fmt.Sprintf(`SELECT %v dt, AVG(conns), 0 FROM ( 
			SELECT date, SUM(b.conns) conns, ip
				FROM %v a, %v_clients b WHERE a.id = b.id %v GROUP by date ORDER BY date, ip
			) GROUP BY dt`, substr, hatchetName, hatchetName, durcond)
	} else if chartType == "total" {
		query = fmt.Sprintf(`SELECT b.ip, SUM(b.accepted), SUM(b.ended)
			FROM %v a, %v_clients b WHERE a.id = b.id %v GROUP by ip ORDER BY accepted DESC;`, hatchetName, hatchetName, durcond)
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
		var doc RemoteClient
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

// GetOpsCounts returns opened connection counts
func (ptr *SQLite3DB) GetOpsCounts(duration string) ([]NameValue, error) {
	docs := []NameValue{}
	var durcond string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
	}
	query := fmt.Sprintf(`SELECT op, COUNT(op) counts
		FROM %v WHERE op != '' %v GROUP by op ORDER BY counts DESC;`, ptr.hatchetName, durcond)
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

// GetReslenByIP returns total response length by ip
func (ptr *SQLite3DB) GetReslenByIP(ip string, duration string) ([]NameValue, error) {
	hatchetName := ptr.hatchetName
	docs := []NameValue{}
	var query, durcond, ipcond string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND a.date BETWEEN '%v' AND '%v'", toks[0], toks[1])
	}
	if ip != "" {
		ipcond = fmt.Sprintf("AND b.ip = '%v'", ip)
		query = fmt.Sprintf(`SELECT a.context, SUM(a.reslen) reslen FROM %v a, %v_clients b
				WHERE reslen > 0 AND a.context = b.context %v %v GROUP by a.context ORDER BY reslen DESC;`,
			hatchetName, hatchetName, ipcond, durcond)
	} else {
		query = fmt.Sprintf(`SELECT b.ip, SUM(a.reslen) reslen FROM %v a, %v_clients b
				WHERE reslen > 0 AND a.context = b.context %v GROUP by b.ip ORDER BY reslen DESC;`,
			hatchetName, hatchetName, durcond)
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

// GetReslenByNamespace returns total response length by ns
func (ptr *SQLite3DB) GetReslenByNamespace(ns string, duration string) ([]NameValue, error) {
	hatchetName := ptr.hatchetName
	docs := []NameValue{}
	var query, durcond, nscond string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
	}
	if ns != "" {
		nscond = fmt.Sprintf("AND ns = '%v'", ns)
		query = fmt.Sprintf(`SELECT ns, SUM(reslen) reslen FROM %v WHERE op != "" AND reslen > 0 %v %v GROUP by ns ORDER BY reslen DESC;`,
			hatchetName, nscond, durcond)
	} else {
		query = fmt.Sprintf(`SELECT ns, SUM(reslen) reslen FROM %v WHERE op != "" AND reslen > 0 %v GROUP by ns ORDER BY reslen DESC;`,
			hatchetName, durcond)
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
