// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"fmt"
	"strings"

	"encoding/json"
	"net/http"
)

// handler responds to API calls
func handler(w http.ResponseWriter, r *http.Request) {
	/** APIs
	 * /tables/{table}/logs
	 * /tables/{table}/logs/slowops
	 * /tables/{table}/stats/slowops
	 *
	 * /api/hatchet/v1.0/tables/{table}/logs/slowops
	 * /api/hatchet/v1.0/tables/{table}/stats/slowops
	 */
	apiPrefix := "/api/hatchet/v1.0/tables/"
	htmlPrefix := "/tables/"
	if r.URL.Path == "/" {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1,
			"URLs": []string{"/tables/{table}/stats/slowops", "/tables/{table}/logs/slowops"}})
		return
	} else if strings.HasPrefix(r.URL.Path, htmlPrefix) {
		tokens := strings.Split(r.URL.Path[len(htmlPrefix):], "/")
		if len(tokens) == 3 {
			tableName := tokens[0]
			category := tokens[1]
			attr := tokens[2]
			if attr == "slowops" && category == "stats" {
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
			} else if attr == "slowops" && category == "logs" {
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
			}
		} else if len(tokens) == 2 && tokens[1] == "logs" {
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
	} else if strings.HasPrefix(r.URL.Path, apiPrefix) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		tokens := strings.Split(r.URL.Path[len(apiPrefix):], "/")
		if len(tokens) == 3 {
			tableName := tokens[0]
			category := tokens[1]
			attr := tokens[2]
			if attr == "slowops" && category == "stats" {
				orderBy := r.URL.Query().Get("orderBy")
				b, err := getSlowOpsStats(tableName, orderBy)
				if err != nil {
					json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				} else {
					w.Write(b)
				}
				return
			} else if attr == "slowops" && category == "logs" {
				topN := ToInt(r.URL.Query().Get("topN"))
				b, err := getSlowLogs(tableName, topN)
				if err != nil {
					json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				} else {
					w.Write(b)
				}
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"table": tableName, "data_type": attr})
		} else if len(tokens) == 2 && tokens[1] == "logs" {
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
			var b []byte
			if b, err = json.Marshal(logs); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			w.Write(b)
			return
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1, "message": "hello hatchet!"})
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

func getSlowLogs(tableName string, topN int) ([]byte, error) {
	if topN == 0 {
		topN = 25
	}
	logstrs, err := getSlowestLogs(tableName, topN)
	if err != nil {
		return nil, err
	}
	return json.Marshal(logstrs)
}
