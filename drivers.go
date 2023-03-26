/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * drivers.go
 */

package hatchet

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var DriversIns *map[string]interface{}

// GetDrivers returns *map[string]interface{} instance
func GetDrivers() *map[string]interface{} {
	if DriversIns == nil {
		filename := "drivers.json"
		data, err := os.ReadFile(filename)
		if err != nil {
			url := "https://raw.githubusercontent.com/simagix/hatchet/main/drivers.json"
			log.Println("download driver manifest from", url)
			resp, err := http.Get(url)
			if err != nil {
				return nil
			}
			defer resp.Body.Close()

			if data, err = io.ReadAll(resp.Body); err != nil {
				return nil
			}
			fname := "drivers.temp"
			log.Println("write driver manifest to", fname)
			_ = os.WriteFile(fname, data, 0644)
		}

		// Parse JSON data into a map
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil
		}

		DriversIns = &m
	}
	return DriversIns
}

func CheckDriverCompatibility(mongo string, driver string, version string) error {
	mongo = strings.TrimPrefix(mongo, "v")
	toks := strings.Split(mongo, ".")
	mongo = strings.Join(toks[:2], ".")
	version = strings.TrimPrefix(version, "v")
	toks = strings.Split(version, ".")
	version = strings.Join(toks[:2], ".")

	drivers := GetDrivers()
	if drivers == nil {
		return fmt.Errorf("missing drivers info")
	}
	driverData, ok := (*drivers)[driver].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing %v driver info", driver)
	}
	driverVerData, ok := driverData[mongo].([]interface{})
	if !ok || len(driverVerData) < 1 {
		return fmt.Errorf(`missing "%v" driver info for MongoDB v%v`, driver, mongo)
	}
	if compareVersions(version, driverVerData[0].(string)) < 0 {
		return fmt.Errorf("MongoDB v%v requires a minimum version of v%v", mongo, driverVerData[0])
	}

	return nil
}

func compareVersions(v1 string, v2 string) int {
	// Split the version strings into their respective integers
	v1Split := strings.Split(v1, ".")
	v2Split := strings.Split(v2, ".")

	a1, _ := strconv.Atoi(v1Split[0])
	b1, _ := strconv.Atoi(v1Split[1])

	a2, _ := strconv.Atoi(v2Split[0])
	b2, _ := strconv.Atoi(v2Split[1])

	// Compare the integers
	if a1 < a2 || (a1 == a2 && b1 < b2) {
		return -1
	} else if a1 > a2 || (a1 == a2 && b1 > b2) {
		return 1
	} else {
		return 0
	}
}
