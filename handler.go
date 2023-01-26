// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"log"
	"net/http"
)

// handler responds to API calls
func handler(w http.ResponseWriter, r *http.Request) {
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
