/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * sqlite3_audit.go
 */

package hatchet

import (
	"fmt"
	"log"
)

func (ptr *SQLite3DB) GetAuditData() (map[string][]NameValues, error) {
	var err error
	db := ptr.db
	data := map[string][]NameValues{}
	// get max connection counts
	query := fmt.Sprintf(`SELECT MAX(conns) FROM %v_clients;`, ptr.hatchetName)
	if ptr.verbose {
		log.Println(query)
	}
	rows, err := db.Query(query)
	category := "stats"
	if err == nil && rows.Next() {
		var doc NameValues
		var value int
		_ = rows.Scan(&value)
		doc.Name = "maxConns"
		doc.Values = append(doc.Values, value)
		if value > 0 {
			data[category] = append(data[category], doc)
		}
		rows.Close()
	}

	// get max operation time
	query = fmt.Sprintf(`SELECT IFNULL(MAX(max_ms), 0), IFNULL(SUM(count), 0), IFNULL(SUM(total_ms), 0) FROM %v_ops;`, ptr.hatchetName)
	if ptr.verbose {
		log.Println(query)
	}
	rows, err = db.Query(query)
	if err == nil && rows.Next() {
		var maxMilli, count, totalMilli int
		if err = rows.Scan(&maxMilli, &count, &totalMilli); err != nil {
			return data, err
		}
		if count > 0 {
			data[category] = append(data[category], NameValues{"maxMilli", []int{maxMilli}})
			data[category] = append(data[category], NameValues{"avgMilli", []int{totalMilli / count}})
			data[category] = append(data[category], NameValues{"totalMilli", []int{totalMilli}})
		}
		rows.Close()
	}

	// get max operation time
	category = "collscan"
	query = fmt.Sprintf(`SELECT IFNULL(MAX(max_ms), 0), IFNULL(SUM(count), 0), IFNULL(SUM(total_ms), 0) FROM %v_ops WHERE _index = 'COLLSCAN';`, ptr.hatchetName)
	if ptr.verbose {
		log.Println(query)
	}
	rows, err = db.Query(query)
	if err == nil && rows.Next() {
		var maxMilli, count, totalMilli int
		if err = rows.Scan(&maxMilli, &count, &totalMilli); err != nil {
			return data, err
		}
		if count > 0 {
			data[category] = append(data[category], NameValues{"count", []int{count}})
			data[category] = append(data[category], NameValues{"maxMilli", []int{maxMilli}})
			data[category] = append(data[category], NameValues{"totalMilli", []int{totalMilli}})
		}
		rows.Close()
	}

	// get audit data
	query = fmt.Sprintf(`SELECT type, name, value FROM %v_audit
		WHERE type IN ('exception', 'failed', 'op', 'duration') ORDER BY type, value DESC;`, ptr.hatchetName)
	if ptr.verbose {
		log.Println(query)
	}
	rows, err = db.Query(query)
	for err == nil && rows.Next() {
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
	if rows != nil {
		rows.Close()
	}

	category = "ip"
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
	if rows != nil {
		rows.Close()
	}

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
	if rows != nil {
		rows.Close()
	}
	return data, err
}
