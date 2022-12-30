// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"net/http"
)

// htmlHandler responds to API calls
func handler(w http.ResponseWriter, r *http.Request) {
	tables, err := getTables()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	templ, err := GetTablesTemplate()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	doc := map[string]interface{}{"Tables": tables}
	if err = templ.Execute(w, doc); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
}
