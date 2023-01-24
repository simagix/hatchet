// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// apiHandler responds to API calls
func apiHandler(w http.ResponseWriter, r *http.Request) {
	/** APIs
	 * /api/hatchet/v1.0/tables/{table}/logs
	 * /api/hatchet/v1.0/tables/{table}/logs/slowops
	 * /api/hatchet/v1.0/tables/{table}/stats/slowops
	 */
	apiPrefix := "/api/hatchet/v1.0/tables/"
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
		limit := r.URL.Query().Get("limit")
		logs, err := getLogs(tableName, fmt.Sprintf("component=%v", component), fmt.Sprintf("limit=%v", limit),
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
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1, "message": "Hello Hatchet API!"})
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
