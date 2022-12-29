// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// htmlHandler responds to API calls
func htmlHandler(w http.ResponseWriter, r *http.Request) {
	/** APIs
	 * /tables/{table}
	 * /tables/{table}/charts/slowops
	 * /tables/{table}/logs
	 * /tables/{table}/logs/slowops
	 * /tables/{table}/stats/slowops
	 */
	htmlPrefix := "/tables/"
	tokens := strings.Split(r.URL.Path[len(htmlPrefix):], "/")
	var tableName, category, attr string
	for i, token := range tokens {
		if i == 0 {
			tableName = token
		} else if i == 1 {
			category = token
		} else if i == 2 {
			attr = token
		}
	}

	if tableName == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": "missing table name"})
		return
	} else if category == "" && attr == "" { // default to /table/{table}/stats/slowops
		category = "stats"
		attr = "slowops"
	}

	if category == "charts" && attr == "slowops" {
		docs, err := getOpCounts(tableName)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetOpCountsTemplate()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		doc := map[string]interface{}{"Table": tableName, "OpCounts": docs}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	} else if category == "logs" && attr == "slowops" {
		topN := ToInt(r.URL.Query().Get("topN"))
		if topN == 0 {
			topN = 25
		}
		logstrs, err := getSlowestLogs(tableName, topN)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetLogsTemplate()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		doc := map[string]interface{}{"Table": tableName, "Logs": logstrs}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	} else if category == "stats" && attr == "slowops" {
		collscan := false
		if r.URL.Query().Get("COLLSCAN") == "true" {
			collscan = true
		}
		orderBy := r.URL.Query().Get("orderBy")
		order := "DESC"
		if orderBy == "" {
			orderBy = "avg_ms"
		} else if orderBy == "index" || orderBy == "_index" {
			orderBy = "_index"
		}
		ops, err := getSlowOps(tableName, orderBy, order, collscan)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetStatsTemplate()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		doc := map[string]interface{}{"Table": tableName, "Ops": ops}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	} else if category == "logs" && attr == "" {
		tableName := tokens[0]
		component := r.URL.Query().Get("component")
		context := r.URL.Query().Get("context")
		severity := r.URL.Query().Get("severity")
		duration := r.URL.Query().Get("duration")
		logs, err := getLogs(tableName, fmt.Sprintf("component=%v", component),
			fmt.Sprintf("context=%v", context), fmt.Sprintf("severity=%v", severity), fmt.Sprintf("duration=%v", duration))
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetLegacyLogTemplate()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		doc := map[string]interface{}{"Table": tableName, "Logs": logs}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	}
}

func getSlowOpsStats(tableName string, orderBy string) ([]byte, error) {
	if orderBy == "" {
		orderBy = "avg_ms"
	}
	ops, err := getSlowOps(tableName, orderBy, "DESC", false)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ops)
}
