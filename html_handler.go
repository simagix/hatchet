// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	 * /tables/{table}
	 * /tables/{table}/charts/slowops
	 * /tables/{table}/logs
	 * /tables/{table}/logs/slowops
	 * /tables/{table}/stats/slowops
	 */
	if GetLogv2().verbose {
		log.Println(r.URL.Path)
	}
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
	summary, start, end := getTableSummary(tableName)
	duration := r.URL.Query().Get("duration")
	download := r.URL.Query().Get("download")
	if duration != "" {
		start, end = getStartEndDates(duration)
	}
	start = strings.Trim(start, "Z")
	end = strings.Trim(end, "Z")

	if category == "charts" && attr == "ops" {
		chartType := "ops"
		docs, err := getAverageOpTime(tableName, duration)
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
		doc := map[string]interface{}{"Table": tableName, "OpCounts": docs, "Chart": charts[chartType],
			"Type": chartType, "Summary": summary, "Start": start, "End": end, "VAxisLabel": "seconds"}
		if err = templ.Execute(w, doc); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		return
	} else if category == "charts" && attr == "slowops" {
		chartType := r.URL.Query().Get("type")
		if GetLogv2().verbose {
			log.Println("type", chartType, "duration", duration)
		}
		if chartType == "" || chartType == "stats" {
			chartType = "slowops"
			docs, err := getSlowOpsCounts(tableName, duration)
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
			doc := map[string]interface{}{"Table": tableName, "OpCounts": docs, "Chart": charts[chartType],
				"Summary": summary, "Start": start, "End": end, "VAxisLabel": "count"}
			if err = templ.Execute(w, doc); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			return
		} else if chartType == "counts" {
			chartType = "slowops-counts"
			docs, err := getOpsCounts(tableName, duration)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			templ, err := GetChartTemplate(chartType)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			doc := map[string]interface{}{"Table": tableName, "NameValues": docs, "Chart": charts[chartType],
				"Summary": summary, "Start": start, "End": end}
			if err = templ.Execute(w, doc); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			return
		}
	} else if category == "charts" && attr == "connections" {
		chartType := r.URL.Query().Get("type")
		if GetLogv2().verbose {
			log.Println("type", chartType, "duration", duration)
		}
		if chartType == "" || chartType == "accepted" {
			chartType = "connections-accepted"
			docs, err := getAcceptedConnsCounts(tableName, duration)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			templ, err := GetChartTemplate(chartType)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			doc := map[string]interface{}{"Table": tableName, "NameValues": docs, "Chart": charts[chartType],
				"Summary": summary, "Start": start, "End": end}
			if err = templ.Execute(w, doc); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
				return
			}
			return
		} else { // type is time or total
			docs, err := getConnectionStats(tableName, chartType, duration)
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
			if len(docs) == 0 {
				docs = []Remote{{IP: "No data", Accepted: 0, Ended: 0}}
			}
			doc := map[string]interface{}{"Table": tableName, "Remote": docs, "Chart": charts[chartType],
				"Summary": summary, "Start": start, "End": end}
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
		logstrs, err := getSlowestLogs(tableName, topN)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetLogTableTemplate(attr)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		doc := map[string]interface{}{"Table": tableName, "Logs": logstrs, "Summary": summary}
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
		ops, err := getSlowOps(tableName, orderBy, order, collscan)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		templ, err := GetStatsTableTemplate(collscan, orderBy, download)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		doc := map[string]interface{}{"Table": tableName, "Ops": ops, "Summary": summary}
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
		limit := r.URL.Query().Get("limit")
		logs, err := getLogs(tableName, fmt.Sprintf("component=%v", component), fmt.Sprintf("limit=%v", limit),
			fmt.Sprintf("context=%v", context), fmt.Sprintf("severity=%v", severity), fmt.Sprintf("duration=%v", duration))
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
		doc := map[string]interface{}{"Table": tableName, "Logs": logs, "LogLength": len(logs), "Seq": seq,
			"Summary": summary, "Context": context, "Component": component, "Severity": severity}
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

func getStartEndDates(duration string) (string, string) {
	toks := strings.Split(duration, ",")
	if len(toks) == 2 {
		return toks[0], toks[1]
	}
	return "", ""
}
