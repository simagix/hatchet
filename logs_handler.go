// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// LogsHandler responds to charts API calls
func LogsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	/** APIs
	 * /hatchets/{hatchet}/logs/all
	 * /hatchets/{hatchet}/logs/slowops
	 */
	hatchetName := params.ByName("hatchet")
	attr := params.ByName("attr")
	dbase, err := GetDatabase(hatchetName)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	defer dbase.Close()
	if dbase.GetVerbose() {
		log.Println("LogsHandler", r.URL.Path, hatchetName, attr)
	}
	info := dbase.GetHatchetInfo()
	summary := GetHatchetSummary(info)
	duration := r.URL.Query().Get("duration")

	if attr == "all" {
		var hasMore bool
		component := r.URL.Query().Get("component")
		context := r.URL.Query().Get("context")
		severity := r.URL.Query().Get("severity")
		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = fmt.Sprintf("%v", LIMIT)
		}
		offset, nlimit := GetOffsetLimit(limit)
		logs, err := dbase.GetLogs(fmt.Sprintf("component=%v", component), fmt.Sprintf("limit=%v", limit),
			fmt.Sprintf("context=%v", context), fmt.Sprintf("severity=%v", severity),
			fmt.Sprintf("duration=%v", duration))
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetLogTableTemplate(attr)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		toks := strings.Split(limit, ",")
		seq := 1
		if len(toks) > 1 {
			seq = ToInt(toks[0]) + 1
		}
		hasMore = len(logs) > nlimit
		if hasMore {
			logs = logs[:len(logs)-1]
		}
		limit = fmt.Sprintf("%v,%v", offset+nlimit, nlimit)
		url := fmt.Sprintf("%v?component=%v&context=%v&severity=%v&duration=%v&limit=%v", r.URL.Path,
			component, context, severity, duration, limit)
		doc := map[string]interface{}{"Hatchet": hatchetName, "Logs": logs, "Seq": seq,
			"Summary": summary, "Context": context, "Component": component, "Severity": severity,
			"HasMore": hasMore, "URL": url}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	} else if attr == "slowops" {
		topN := ToInt(r.URL.Query().Get("topN"))
		if topN == 0 {
			topN = TOP_N
		}
		logstrs, err := dbase.GetSlowestLogs(topN)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetLogTableTemplate(attr)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		doc := map[string]interface{}{"Hatchet": hatchetName, "Logs": logstrs, "Summary": summary}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	}
}
