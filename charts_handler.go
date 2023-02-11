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

// ChartsHandler responds to charts API calls
func ChartsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	/** APIs
	 * /hatchets/{hatchet}/charts/slowops
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

	if attr == "ops" {
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
	} else if attr == "slowops" {
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
	} else if attr == "connections" {
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
