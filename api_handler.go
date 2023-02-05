// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// apiHandler responds to API calls
func apiHandler(w http.ResponseWriter, r *http.Request) {
	/** APIs
	 * /api/hatchet/v1.0/hatchets/{hatchet}/logs
	 * /api/hatchet/v1.0/hatchets/{hatchet}/logs/slowops
	 * /api/hatchet/v1.0/hatchets/{hatchet}/stats/slowops
	 */
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	tokens := strings.Split(r.URL.Path[len(REST_API_PREFIX):], "/")
	var hatchetName string
	if len(tokens) > 0 {
		hatchetName = tokens[0]
	}
	dbase, err := GetDatabase(hatchetName)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	defer dbase.Close()
	if dbase.GetVerbose() {
		log.Println(r.URL.Path)
	}

	if len(tokens) == 3 {
		category := tokens[1]
		attr := tokens[2]
		if attr == "slowops" && category == "stats" {
			orderBy := r.URL.Query().Get("orderBy")
			if orderBy == "" {
				orderBy = "avg_ms"
			}
			ops, err := dbase.GetSlowOps(orderBy, "DESC", false)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			}
			doc := map[string]interface{}{"hatchet": hatchetName, "has_more": false, "offset": 0, "limit": len(ops), "ops": ops}
			b, err := json.Marshal(doc)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			} else {
				w.Write(b)
			}
			return
		} else if attr == "slowops" && category == "logs" {
			topN := ToInt(r.URL.Query().Get("topN"))
			if topN == 0 {
				topN = TOP_N
			}
			logs, err := dbase.GetSlowestLogs(topN)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			}
			doc := map[string]interface{}{"hatchet": hatchetName, "has_more": false, "offset": 0, "limit": len(logs), "logs": logs}
			b, err := json.Marshal(doc)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			} else {
				w.Write(b)
			}
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"hatchet": hatchetName, "data_type": attr})
	} else if len(tokens) == 2 && tokens[1] == "logs" {
		var hasMore bool
		component := r.URL.Query().Get("component")
		context := r.URL.Query().Get("context")
		severity := r.URL.Query().Get("severity")
		duration := r.URL.Query().Get("duration")
		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = fmt.Sprintf("%v", LIMIT)
		}
		offset, nlimit := GetOffsetLimit(limit)
		logs, err := dbase.GetLogs(fmt.Sprintf("component=%v", component), fmt.Sprintf("limit=%v", limit),
			fmt.Sprintf("context=%v", context), fmt.Sprintf("severity=%v", severity), fmt.Sprintf("duration=%v", duration))
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		hasMore = len(logs) > nlimit
		if hasMore {
			logs = logs[:len(logs)-1]
		}
		var b []byte
		doc := map[string]interface{}{"hatchet": hatchetName, "has_more": hasMore, "offset": offset, "limit": len(logs), "logs": logs}
		if b, err = json.Marshal(doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		w.Write(b)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1, "message": "Hello Hatchet API!"})
}
