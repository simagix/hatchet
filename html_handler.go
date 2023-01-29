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
	HTML_URL_PREFIX = "/hatchets/"
)

type Chart struct {
	Index int
	Label string
	Title string
	URL   string
}

var charts = map[string]Chart{
	"instruction":          {0, "select a chart", "", ""},
	"ops":                  {1, "avg op time", "Average Operation Time", "/ops?type=stats"},
	"slowops":              {2, "ops stats", "Slow Operation Counts", "/slowops?type=stats"},
	"slowops-counts":       {3, "ops counts", "Operation Counts", "/slowops?type=counts"},
	"connections-accepted": {4, "conns accepted", "Accepted Connections", "/connections?type=accepted"},
	"connections-time":     {5, "conns by time", "Accepted vs Ended Connections", "/connections?type=time"},
	"connections-total":    {6, "conns by total", "Accepted vs Ended Connections by IP", "/connections?type=total"},
}

// htmlHandler responds to API calls
func htmlHandler(w http.ResponseWriter, r *http.Request) {
	/** APIs
	 * /hatchets/{hatchet}
	 * /hatchets/{hatchet}/charts/slowops
	 * /hatchets/{hatchet}/logs
	 * /hatchets/{hatchet}/logs/slowops
	 * /hatchets/{hatchet}/stats/slowops
	 */
	tokens := strings.Split(r.URL.Path[len(HTML_URL_PREFIX):], "/")
	var hatchetName, category, attr string
	for i, token := range tokens {
		if i == 0 {
			hatchetName = token
		} else if i == 1 {
			category = token
		} else if i == 2 {
			attr = token
		}
	}
	if hatchetName == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": "missing hatchet name"})
		return
	} else if category == "" && attr == "" { // default to /hatchets/{hatchet}/stats/slowops
		category = "stats"
		attr = "slowops"
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
	info := dbase.GetHatchetInfo()
	summary := GetHatchetSummary(info)
	start, end := getStartEndDates(fmt.Sprintf("%v,%v", info.Start, info.End))
	duration := r.URL.Query().Get("duration")
	download := r.URL.Query().Get("download")
	if duration != "" {
		start, end = getStartEndDates(duration)
	}

	if category == "charts" && attr == "ops" {
		chartType := "ops"
		docs, err := dbase.GetAverageOpTime(duration)
		if len(docs) > 0 {
			start = docs[0].Date
			end = docs[len(docs)-1].Date
		}
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetChartTemplate(chartType)
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
		return
	} else if category == "charts" && attr == "slowops" {
		chartType := r.URL.Query().Get("type")
		if dbase.GetVerbose() {
			log.Println("type", chartType, "duration", duration)
		}
		if chartType == "" || chartType == "stats" {
			chartType = "slowops"
			docs, err := dbase.GetSlowOpsCounts(duration)
			if len(docs) > 0 {
				start = docs[0].Date
				end = docs[len(docs)-1].Date
			}
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			templ, err := GetChartTemplate(chartType)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			doc := map[string]interface{}{"Hatchet": hatchetName, "OpCounts": docs, "Chart": charts[chartType],
				"Type": chartType, "Summary": summary, "Start": start, "End": end, "VAxisLabel": "count"}
			if err = templ.Execute(w, doc); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			return
		} else if chartType == "counts" {
			chartType = "slowops-counts"
			docs, err := dbase.GetOpsCounts(duration)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			templ, err := GetChartTemplate(chartType)
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
	} else if category == "charts" && attr == "connections" {
		chartType := r.URL.Query().Get("type")
		if dbase.GetVerbose() {
			log.Println("type", chartType, "duration", duration)
		}
		if chartType == "" || chartType == "accepted" {
			chartType = "connections-accepted"
			docs, err := dbase.GetAcceptedConnsCounts(duration)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			templ, err := GetChartTemplate(chartType)
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
			templ, err := GetChartTemplate(chartType)
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
	} else if category == "logs" && attr == "slowops" {
		topN := ToInt(r.URL.Query().Get("topN"))
		if topN == 0 {
			topN = 25
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
	} else if category == "stats" && attr == "slowops" {
		collscan := false
		if r.URL.Query().Get("COLLSCAN") == "true" {
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
	} else if category == "logs" && attr == "" {
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
