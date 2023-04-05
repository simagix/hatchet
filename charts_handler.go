/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * charts_handler.go
 */

package hatchet

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

const (
	BAR_CHART    = "bar_chart"
	BUBBLE_CHART = "bubble_chart"
	PIE_CHART    = "pie_chart"

	T_OPS            = "ops"
	T_RESLEN_UP      = "reslen-ip"
	T_OPS_COUNTS     = "ops-counts"
	T_CONNS_ACCEPTED = "connections-accepted"
	T_CONNS_TIME     = "connections-time"
	T_CONNS_TOTAL    = "connections-total"
	T_RESLEN_NS      = "reslen-ns"
)

type Chart struct {
	Index int
	Title string
	Descr string
	URL   string
}

var charts = map[string]Chart{
	"instruction": {0, "select a chart", "", ""},
	T_OPS: {1, "Average Operation Time",
		"Display average operations time over a period of time", "/ops?type=stats"},
	T_OPS_COUNTS: {2, "Operation Counts",
		"Display total counts of operations", "/ops?type=counts"},
	T_CONNS_TIME: {3, "Average Connections",
		"Display accepted vs ended connections over a period of time", "/connections?type=time"},
	T_CONNS_ACCEPTED: {4, "Accepted Connections",
		"Display accepted connections from clients", "/connections?type=accepted"},
	T_CONNS_TOTAL: {5, "Accepted & Ended from IPs",
		"Display accepted vs ended connections by client IPs", "/connections?type=total"},
	T_RESLEN_UP: {6, "Response Length by IPs ",
		"Display total response length by client IPs", "/reslen-ip?ip="},
	T_RESLEN_NS: {7, "Response Length by Namespaces ",
		"Display total response length by namespaces", "/reslen-ns?ns="},
}

// ChartsHandler responds to charts API calls
func ChartsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	/** APIs
	 * /hatchets/{hatchet}/charts/ops
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
		log.Println("ChartsHandler", r.URL.Path, hatchetName, attr)
	}
	info := dbase.GetHatchetInfo()
	summary := GetHatchetSummary(info)
	start, end := getStartEndDates(fmt.Sprintf("%v,%v", info.Start, info.End))
	duration := r.URL.Query().Get("duration")
	if duration != "" {
		start, end = getStartEndDates(duration)
	}

	if attr == T_OPS {
		chartType := r.URL.Query().Get("type")
		op := r.URL.Query().Get("op")
		if chartType == "stats" {
			chartType := T_OPS
			docs, err := dbase.GetAverageOpTime(op, duration)
			if len(docs) > 0 {
				start = docs[0].Date
				end = docs[len(docs)-1].Date
			}
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			templ, err := GetChartTemplate(BUBBLE_CHART)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			doc := map[string]interface{}{"Hatchet": hatchetName, "OpCounts": docs, "Chart": charts[chartType],
				"Type": chartType, "Summary": summary, "Start": start, "End": end, "VAxisLabel": "seconds"}
			if err = templ.Execute(w, doc); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
		} else if chartType == "counts" {
			chartType = T_OPS_COUNTS
			docs, err := dbase.GetOpsCounts(duration)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			templ, err := GetChartTemplate(PIE_CHART)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			doc := map[string]interface{}{"Hatchet": hatchetName, "NameValues": docs, "Chart": charts[chartType],
				"Type": chartType, "Summary": summary, "Start": start, "End": end}
			if err = templ.Execute(w, doc); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			return
		}
		return
	} else if attr == "connections" {
		chartType := r.URL.Query().Get("type")
		if dbase.GetVerbose() {
			log.Println("type", chartType, "duration", duration)
		}
		if chartType == "" || chartType == "accepted" {
			chartType = T_CONNS_ACCEPTED
			docs, err := dbase.GetAcceptedConnsCounts(duration)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			templ, err := GetChartTemplate(PIE_CHART)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			doc := map[string]interface{}{"Hatchet": hatchetName, "NameValues": docs, "Chart": charts[chartType],
				"Type": chartType, "Summary": summary, "Start": start, "End": end}
			if err = templ.Execute(w, doc); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			return
		} else { // type is time or total
			docs, err := dbase.GetConnectionStats(chartType, duration)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			chartType = "connections-" + chartType
			templ, err := GetChartTemplate(BAR_CHART)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			doc := map[string]interface{}{"Hatchet": hatchetName, "Remote": docs, "Chart": charts[chartType],
				"Type": chartType, "Summary": summary, "Start": start, "End": end}
			if err = templ.Execute(w, doc); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			return
		}
	} else if attr == T_RESLEN_UP {
		ip := r.URL.Query().Get("ip")
		chartType := attr
		if dbase.GetVerbose() {
			log.Println("type", chartType, "duration", duration)
		}
		docs, err := dbase.GetReslenByIP(ip, duration)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetChartTemplate(PIE_CHART)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		chart := charts[chartType]
		if ip != "" {
			chart.Title += fmt.Sprintf(" (%v)", ip)
		}
		doc := map[string]interface{}{"Hatchet": hatchetName, "NameValues": docs, "Chart": chart,
			"Type": chartType, "Summary": summary, "Start": start, "End": end}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	} else if attr == T_RESLEN_NS {
		ns := r.URL.Query().Get("ns")
		chartType := attr
		if dbase.GetVerbose() {
			log.Println("type", chartType, "duration", duration)
		}
		docs, err := dbase.GetReslenByNamespace(ns, duration)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetChartTemplate(PIE_CHART)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		chart := charts[chartType]
		if ns != "" {
			chart.Title += fmt.Sprintf(" (%v)", ns)
		}
		doc := map[string]interface{}{"Hatchet": hatchetName, "NameValues": docs, "Chart": chart,
			"Type": chartType, "Summary": summary, "Start": start, "End": end}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	}
}

func getStartEndDates(duration string) (string, string) {
	var start, end string
	toks := strings.Split(duration, ",")
	if len(toks) == 2 {
		if len(toks[0]) >= 16 {
			start = toks[0][:16]
		}
		if len(toks[1]) >= 16 {
			end = toks[1][:16]
		}
	}
	return start, end
}
