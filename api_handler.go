// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	TOPN = 25
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
	dbase, err := GetDatabase()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	defer dbase.Close()
	if dbase.GetVerbose() {
		log.Println(r.URL.Path)
	}

	if len(tokens) == 3 {
		tableName := tokens[0]
		category := tokens[1]
		attr := tokens[2]
		if attr == "slowops" && category == "stats" {
			orderBy := r.URL.Query().Get("orderBy")
			if orderBy == "" {
				orderBy = "avg_ms"
			}
			ops, err := dbase.GetSlowOps(tableName, orderBy, "DESC", false)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			}
			b, err := json.Marshal(ops)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			} else {
				w.Write(b)
			}
			return
		} else if attr == "slowops" && category == "logs" {
			topN := ToInt(r.URL.Query().Get("topN"))
			if topN == 0 {
				topN = TOPN
			}
			logstrs, err := dbase.GetSlowestLogs(tableName, topN)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			}
			b, err := json.Marshal(logstrs)
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
		logs, err := dbase.GetLogs(tableName, fmt.Sprintf("component=%v", component), fmt.Sprintf("limit=%v", limit),
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
