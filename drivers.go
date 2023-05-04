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
	"strings"
)

var driversIns *map[string]interface{}

// GetDrivers returns *map[string]interface{} instance
func GetDrivers() *map[string]interface{} {
	if driversIns == nil {
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
		driversIns = &m
	}
	return driversIns
}

func GetDriverVersions(mongo string, driver string) ([]interface{}, error) {
	var versions []interface{}
	if mongo == "" {
		return versions, fmt.Errorf("missing MongoDB version")
	} else if driver == "" {
		return versions, fmt.Errorf("missing driver info")
	}
	parts := strings.Split(driver, "|")
	if len(parts) == 0 {
		return versions, fmt.Errorf("missing driver info")
	}
	version := parts[0]
	mongo = strings.TrimPrefix(mongo, "v")
	toks := strings.Split(mongo, ".")
	mongo = strings.Join(toks[:2], ".")

	drivers := GetDrivers()
	if drivers == nil {
		return versions, fmt.Errorf("missing driver data")
	}
	driverData, ok := (*drivers)[mongo].(map[string]interface{})
	if !ok {
		return versions, fmt.Errorf("missing MongoDB v%v driver data", mongo)
	}
	versions, ok = driverData[version].([]interface{})
	if !ok || len(versions) < 1 {
		return versions, fmt.Errorf(`missing MongoDB v%v driver "%v" data`, mongo, version)
	}
	return versions, nil
}

func CheckDriverCompatibility(mongo string, driver string, version string) error {
	versions, err := GetDriverVersions(mongo, driver)
	if err != nil {
		return err
	}
	if version == "" {
		return fmt.Errorf("missing driver info")
	}
	version = strings.TrimPrefix(version, "v")
	toks := strings.Split(version, ".")
	version = strings.Join(toks[:2], ".")
	if compareVersions(version, versions[0].(string)) < 0 {
		return fmt.Errorf("MongoDB v%v requires a minimum driver version of v%v", mongo, versions[0])
	}
	return nil
}

func compareVersions(v1 string, v2 string) int {
	// Split the version strings into their respective integers
	v1Split := strings.Split(v1, ".")
	a1 := ToInt(v1Split[0])
	b1 := ToInt(v1Split[1])

	v2Split := strings.Split(v2, ".")
	a2 := ToInt(v2Split[0])
	b2 := ToInt(v2Split[1])

	// Compare the integers
	if a1 < a2 || (a1 == a2 && b1 < b2) {
		return -1
	} else if a1 > a2 || (a1 == a2 && b1 > b2) {
		return 1
	} else {
		return 0
	}
}
