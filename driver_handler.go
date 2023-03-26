/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * driver_handler.go
 */

package hatchet

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// DriverHandler responds to API calls
func DriverHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	/** APIs
	 * /api/hatchet/v1.0/mongodb/{mongo}/drivers/{driver}?compatibleWith={version}
	 */
	w.Header().Set("Content-Type", "application/json")
	driver := params.ByName("driver")
	mongo := params.ByName("mongo")
	version := r.URL.Query().Get("compatibleWith")

	if version == "" {
		versions, err := GetDriverVersions(mongo, driver)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1, "MongoDB": mongo,
			"driver": map[string]interface{}{"name": driver, "versions": versions}})
		return
	} else if err := CheckDriverCompatibility(mongo, driver, version); err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": 0, "error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1})
}
