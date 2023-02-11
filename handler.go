// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	HTML_API_PREFIX = "/hatchets/"
	REST_API_PREFIX = "/api/hatchet/v1.0/hatchets/"
)

// Handler responds to API calls
func Handler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	dbase, err := GetDatabase(GetLogv2().hatchetName) // main page
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	defer dbase.Close()
	if dbase.GetVerbose() {
		log.Println(r.URL.Path)
	}
	hatchets, err := dbase.GetHatchetNames()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	templ, err := GetTablesTemplate()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	doc := map[string]interface{}{"Hatchets": hatchets, "Version": GetLogv2().version}
	if err = templ.Execute(w, doc); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
}
