/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * sqlite3_query.go
 */

package hatchet

import (
	"fmt"
	"log"
	"strings"
)

type OpCount struct {
	Date      string
	Count     int
	Milli     float64
	Op        string
	Namespace string
	Filter    string
}

func (ptr *SQLite3DB) GetSlowOps(orderBy string, order string, collscan bool) ([]OpStat, error) {
	ops := []OpStat{}
	db := ptr.db
	query := fmt.Sprintf(`SELECT op, count, avg_ms, max_ms,
			total_ms, ns, _index "index", reslen, filter "query pattern"
			FROM %v_ops ORDER BY %v %v`, ptr.hatchetName, orderBy, order)
	if collscan {
		query = fmt.Sprintf(`SELECT op, count, avg_ms, max_ms,
				total_ms, ns, _index "index", reslen, filter "query pattern"
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
		substr = GetDateSubString(toks[0], toks[1])
	} else {
		info := ptr.GetHatchetInfo()
		substr = GetDateSubString(info.Start, info.End)
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
	defer rows.Close()
	if rows.Next() {
		if err = rows.Scan(&info.Name, &info.Version, &info.Module, &info.OS, &info.Arch,
			&info.Start, &info.End); err != nil {
			return info
		}
	}
	return info
}

func (ptr *SQLite3DB) GetHatchetNames() ([]string, error) {
	hatchets := []string{}
	query := "SELECT name, version, module, os, arch FROM hatchet ORDER BY name"
	db := ptr.db
	if ptr.verbose {
		log.Println(query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return hatchets, err
	}
	defer rows.Close()
	for rows.Next() {
		var name, version, module, os, arch string
		if err = rows.Scan(&name, &version, &module, &os, &arch); err != nil {
			return hatchets, err
		}
		hatchets = append(hatchets, name)
	}
	return hatchets, err
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
func (ptr *SQLite3DB) GetConnectionStats(chartType string, duration string) ([]Remote, error) {
	hatchetName := ptr.hatchetName
	docs := []Remote{}
	var query, durcond string
	var substr string
	if duration != "" {
		toks := strings.Split(duration, ",")
		durcond = fmt.Sprintf("AND date BETWEEN '%v' AND '%v'", toks[0], toks[1])
		substr = GetDateSubString(toks[0], toks[1])
	} else {
		info := ptr.GetHatchetInfo()
		substr = GetDateSubString(info.Start, info.End)
	}
	if chartType == "time" {
		query = fmt.Sprintf(`SELECT %v dt, SUM(b.accepted), SUM(b.ended)
			FROM %v a, %v_clients b WHERE a.id = b.id %v GROUP by dt ORDER BY dt;`,
			substr, hatchetName, hatchetName, durcond)
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

func (ptr *SQLite3DB) GetAuditData() (map[string][]NameValues, error) {
	var err error
	data := map[string][]NameValues{}
	query := fmt.Sprintf(`SELECT type, name, value FROM %v_audit
		WHERE type IN ('exception', 'failed', 'op', 'duration') ORDER BY type, value DESC;`, ptr.hatchetName)
	if ptr.verbose {
		log.Println(query)
	}
	db := ptr.db
	rows, err := db.Query(query)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		var category string
		var doc NameValues
		var value int
		if err = rows.Scan(&category, &doc.Name, &value); err != nil {
			return data, err
		}
		doc.Values = append(doc.Values, value)
		if category == "exception" {
			if doc.Name == "E" {
				doc.Name = "Error"
			} else if doc.Name == "F" {
				doc.Name = "Fatal"
			} else if doc.Name == "W" {
				doc.Name = "Warn"
			}
		}
		data[category] = append(data[category], doc)
	}
	rows.Close()

	category := "ip"
	query = fmt.Sprintf(`SELECT a.name ip, a.value count, b.value reslen FROM %v_audit a, %v_audit b WHERE a.type == '%v' AND b.type = 'reslen-ip' AND a.name = b.name ORDER BY reslen DESC;`,
		ptr.hatchetName, ptr.hatchetName, category)
	if ptr.verbose {
		log.Println(query)
	}
	rows, err = db.Query(query)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		var doc NameValues
		var val1, val2 int
		if err = rows.Scan(&doc.Name, &val1, &val2); err != nil {
			return data, err
		}
		doc.Values = append(doc.Values, val1)
		doc.Values = append(doc.Values, val2)
		data[category] = append(data[category], doc)
	}
	rows.Close()

	category = "ns"
	query = fmt.Sprintf(`SELECT a.name ns, a.value count, b.value reslen FROM %v_audit a, %v_audit b WHERE a.type == '%v' AND b.type = 'reslen-ns' AND a.name = b.name ORDER BY reslen DESC;`,
		ptr.hatchetName, ptr.hatchetName, category)
	if ptr.verbose {
		log.Println(query)
	}
	rows, err = db.Query(query)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		var doc NameValues
		var val1, val2 int
		if err = rows.Scan(&doc.Name, &val1, &val2); err != nil {
			return data, err
		}
		doc.Values = append(doc.Values, val1)
		doc.Values = append(doc.Values, val2)
		data[category] = append(data[category], doc)
	}
	defer rows.Close()
	return data, err
}
