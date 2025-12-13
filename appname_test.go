/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * appname_test.go
 */

package hatchet

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestAppNameParsing(t *testing.T) {
	// Test log with appName
	str := `{"t":{"$date":"2021-07-25T09:38:57.078+00:00"},"s":"I", "c":"COMMAND", "id":51803, "ctx":"conn541","msg":"Slow query","attr":{"type":"command","ns":"demo.hatchet","appName":"Keyhole Lib","command":{"aggregate":"hatchet","allowDiskUse":true,"pipeline":[{"$match":{"status":{"$in":["completed","split","splitting"]}}}],"cursor":{},"$db":"demo"},"planSummary":"IXSCAN { status: 1 }","keysExamined":218,"docsExamined":217,"numYields":6,"nreturned":53,"reslen":6117,"durationMillis":530}}`

	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	// Parse attributes
	b, _ := bson.Marshal(doc.Attr)
	bson.Unmarshal(b, &doc.Attributes)

	// Verify appName is parsed correctly
	if doc.Attributes.AppName != "Keyhole Lib" {
		t.Fatalf("expected appName 'Keyhole Lib', but got '%v'", doc.Attributes.AppName)
	}
	t.Logf("Successfully parsed appName: %s", doc.Attributes.AppName)
}

func TestAppNameEmpty(t *testing.T) {
	// Test log without appName
	str := `{"t":{"$date":"2021-07-25T09:38:57.078+00:00"},"s":"I", "c":"COMMAND", "id":51803, "ctx":"conn541","msg":"Slow query","attr":{"type":"command","ns":"demo.hatchet","command":{"find":"hatchet","filter":{},"$db":"demo"},"planSummary":"COLLSCAN","keysExamined":0,"docsExamined":100,"numYields":1,"nreturned":100,"reslen":5000,"durationMillis":100}}`

	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	// Parse attributes
	b, _ := bson.Marshal(doc.Attr)
	bson.Unmarshal(b, &doc.Attributes)

	// Verify appName is empty
	if doc.Attributes.AppName != "" {
		t.Fatalf("expected empty appName, but got '%v'", doc.Attributes.AppName)
	}
	t.Log("Successfully handled log without appName")
}

func TestGetReslenByAppName(t *testing.T) {
	dbfile := "/tmp/test_appname.db"
	hatchetName := "test_appname"

	// Clean up any existing test database
	os.Remove(dbfile)
	defer os.Remove(dbfile)

	// Create SQLite database
	sqlite, err := NewSQLite3DB(dbfile, hatchetName, 2000)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.Close()

	// Begin transaction and create tables
	if err = sqlite.Begin(); err != nil {
		t.Fatal(err)
	}

	// Insert test data with appName - using proper log format with filter field
	testLogs := []struct {
		logStr string
	}{
		{`{"t":{"$date":"2021-07-25T09:38:57.078+00:00"},"s":"I","c":"COMMAND","id":51803,"ctx":"conn1","msg":"Slow query","attr":{"type":"command","ns":"demo.test","appName":"App1","command":{"find":"test","filter":{"x":1},"$db":"demo"},"planSummary":"IXSCAN { x: 1 }","reslen":1000,"durationMillis":100}}`},
		{`{"t":{"$date":"2021-07-25T09:38:58.078+00:00"},"s":"I","c":"COMMAND","id":51803,"ctx":"conn2","msg":"Slow query","attr":{"type":"command","ns":"demo.test","appName":"App1","command":{"find":"test","filter":{"x":1},"$db":"demo"},"planSummary":"IXSCAN { x: 1 }","reslen":2000,"durationMillis":150}}`},
		{`{"t":{"$date":"2021-07-25T09:38:59.078+00:00"},"s":"I","c":"COMMAND","id":51803,"ctx":"conn3","msg":"Slow query","attr":{"type":"command","ns":"demo.test","appName":"App2","command":{"find":"test","filter":{"y":1},"$db":"demo"},"planSummary":"COLLSCAN","reslen":3000,"durationMillis":200}}`},
	}

	for i, test := range testLogs {
		doc := Logv2Info{}
		if err := bson.UnmarshalExtJSON([]byte(test.logStr), false, &doc); err != nil {
			t.Fatalf("Failed to parse test log %d: %v", i, err)
		}
		if err := AddLegacyString(&doc); err != nil {
			t.Fatalf("Failed to add legacy string for log %d: %v", i, err)
		}
		stat, _ := AnalyzeSlowOp(&doc)
		end := getDateTimeStr(doc.Timestamp)
		if err := sqlite.InsertLog(i+1, end, &doc, stat); err != nil {
			t.Fatalf("Failed to insert log %d: %v", i, err)
		}
	}

	// Commit transaction
	if err = sqlite.Commit(); err != nil {
		t.Fatal(err)
	}

	// Create metadata (indexes and audit entries)
	if err = sqlite.CreateMetaData(); err != nil {
		t.Fatal(err)
	}

	// Test GetReslenByAppName
	docs, err := sqlite.GetReslenByAppName("", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(docs) != 2 {
		t.Fatalf("expected 2 appNames, but got %d", len(docs))
	}

	// Verify total reslen for each appName
	appNameReslen := make(map[string]int)
	for _, doc := range docs {
		appNameReslen[doc.Name] = doc.Value
	}

	if appNameReslen["App1"] != 3000 {
		t.Fatalf("expected App1 reslen 3000, but got %d", appNameReslen["App1"])
	}
	if appNameReslen["App2"] != 3000 {
		t.Fatalf("expected App2 reslen 3000, but got %d", appNameReslen["App2"])
	}

	t.Logf("GetReslenByAppName returned %d entries", len(docs))
	for _, doc := range docs {
		t.Logf("  AppName: %s, Reslen: %d", doc.Name, doc.Value)
	}
}

func TestAppNameAuditData(t *testing.T) {
	dbfile := "/tmp/test_appname_audit.db"
	hatchetName := "test_appname_audit"

	// Clean up any existing test database
	os.Remove(dbfile)
	defer os.Remove(dbfile)

	// Create SQLite database
	sqlite, err := NewSQLite3DB(dbfile, hatchetName, 2000)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.Close()

	// Begin transaction and create tables
	if err = sqlite.Begin(); err != nil {
		t.Fatal(err)
	}

	// Insert test data with different appNames - using proper log format
	testLogs := []struct {
		logStr string
	}{
		{`{"t":{"$date":"2021-07-25T09:38:57.078+00:00"},"s":"I","c":"COMMAND","id":51803,"ctx":"conn1","msg":"Slow query","attr":{"type":"command","ns":"demo.test","appName":"MyApp","command":{"find":"test","filter":{"x":1},"$db":"demo"},"planSummary":"IXSCAN { x: 1 }","reslen":1000,"durationMillis":100}}`},
		{`{"t":{"$date":"2021-07-25T09:38:58.078+00:00"},"s":"I","c":"COMMAND","id":51803,"ctx":"conn2","msg":"Slow query","attr":{"type":"command","ns":"demo.test","appName":"MyApp","command":{"find":"test","filter":{"x":1},"$db":"demo"},"planSummary":"IXSCAN { x: 1 }","reslen":2000,"durationMillis":150}}`},
		{`{"t":{"$date":"2021-07-25T09:38:59.078+00:00"},"s":"I","c":"COMMAND","id":51803,"ctx":"conn3","msg":"Slow query","attr":{"type":"command","ns":"demo.test","appName":"OtherApp","command":{"find":"test","filter":{"y":1},"$db":"demo"},"planSummary":"COLLSCAN","reslen":5000,"durationMillis":200}}`},
	}

	for i, test := range testLogs {
		doc := Logv2Info{}
		if err := bson.UnmarshalExtJSON([]byte(test.logStr), false, &doc); err != nil {
			t.Fatalf("Failed to parse test log %d: %v", i, err)
		}
		if err := AddLegacyString(&doc); err != nil {
			t.Fatalf("Failed to add legacy string for log %d: %v", i, err)
		}
		stat, _ := AnalyzeSlowOp(&doc)
		end := getDateTimeStr(doc.Timestamp)
		if err := sqlite.InsertLog(i+1, end, &doc, stat); err != nil {
			t.Fatalf("Failed to insert log %d: %v", i, err)
		}
	}

	// Commit transaction
	if err = sqlite.Commit(); err != nil {
		t.Fatal(err)
	}

	// Create metadata
	if err = sqlite.CreateMetaData(); err != nil {
		t.Fatal(err)
	}

	// Get audit data
	data, err := sqlite.GetAuditData()
	if err != nil {
		t.Fatal(err)
	}

	// Check appname data exists
	appnameData, ok := data["appname"]
	if !ok {
		t.Fatal("appname data not found in audit data")
	}

	if len(appnameData) != 2 {
		t.Fatalf("expected 2 appNames in audit data, but got %d", len(appnameData))
	}

	t.Logf("Audit data contains %d appName entries:", len(appnameData))
	for _, entry := range appnameData {
		t.Logf("  AppName: %s, Count: %v, Reslen: %v", entry.Name, entry.Values[0], entry.Values[1])
	}
}

