// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"strings"

	"encoding/json"
	"net/http"
)

// handler responds to API calls
func handler(w http.ResponseWriter, r *http.Request) {
	/** APIs
	 * /tables/{table}/slowops/summary
	 *
	 * /api/hatchet/v1.0/tables/{table}/slowops/summary
	 * /api/hatchet/v1.0/tables/{table}/slowops/logs
	 */
	apiPrefix := "/api/hatchet/v1.0/"
	htmlPrefix := "/tables/"
	if r.URL.Path == "/" {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1, "URL": "/tables/{table}/slowops/summary"})
		return
	} else if strings.HasPrefix(r.URL.Path, htmlPrefix) {
		tokens := strings.Split(r.URL.Path[1:], "/")
		if len(tokens) == 4 && tokens[0] == "tables" {
			tableName := tokens[1]
			category := tokens[2]
			dataType := tokens[3]
			if category == "slowops" && dataType == "summary" {
				orderBy := r.URL.Query().Get("orderBy")
				if orderBy == "" {
					orderBy = "avg_ms"
				}
				ops, err := getSlowOps(tableName, orderBy, "DESC")
				if err != nil {
					json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
					return
				}
				templ, err := GetSummaryTemplate()
				if err != nil {
					json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
					return
				}
				doc := map[string]interface{}{"Table": tableName, "Ops": ops}
				if err = templ.Execute(w, doc); err != nil {
					json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
					return
				}
			}
			return
		}
	} else if strings.HasPrefix(r.URL.Path, apiPrefix) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		tokens := strings.Split(r.URL.Path[len(apiPrefix):], "/")
		if len(tokens) == 4 && tokens[0] == "tables" {
			tableName := tokens[1]
			category := tokens[2]
			dataType := tokens[3]
			if category == "slowops" && dataType == "summary" {
				orderBy := r.URL.Query().Get("orderBy")
				b, err := getSlowOpsSummary(tableName, orderBy)
				if err != nil {
					json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				} else {
					w.Write(b)
				}
				return
			} else if category == "slowops" && dataType == "logs" {
				topN := ToInt(r.URL.Query().Get("topN"))
				b, err := getSlowLogs(tableName, topN)
				if err != nil {
					json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				} else {
					w.Write(b)
				}
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"table": tableName, "data_type": category})
		}
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1, "message": "hello hatchet!"})
}

func getSlowOpsSummary(tableName string, orderBy string) ([]byte, error) {
	if orderBy == "" {
		orderBy = "avg_ms"
	}
	ops, err := getSlowOps(tableName, orderBy, "DESC")
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
