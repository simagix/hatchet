// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// StatsHandler responds to API calls
func StatsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	/** APIs
	 * /hatchets/{hatchet}/stats/slowops
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
		log.Println("StatsHandler", r.URL.Path, hatchetName, attr)
	}
	info := dbase.GetHatchetInfo()
	summary := GetHatchetSummary(info)
	download := r.URL.Query().Get("download")

	if attr == "slowops" {
		collscan := false
		if r.URL.Query().Get(COLLSCAN) == "true" {
			collscan = true
		}
		var order, orderBy string
		orderBy = r.URL.Query().Get("orderBy")
		if orderBy == "" {
			orderBy = "avg_ms"
		} else if orderBy == "index" || orderBy == "_index" {
			orderBy = "_index"
		}
		order = r.URL.Query().Get("order")
		if order == "" {
			if orderBy == "op" || orderBy == "ns" {
				order = "ASC"
			} else {
				order = "DESC"
			}
		}
		ops, err := dbase.GetSlowOps(orderBy, order, collscan)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetStatsTableTemplate(collscan, orderBy, download)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		doc := map[string]interface{}{"Hatchet": hatchetName, "Ops": ops, "Summary": summary}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	}
}
