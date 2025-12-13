/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * benchmark_test.go
 */

package hatchet

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// Sample slow query log for benchmarking
var benchmarkLogStr = `{"t":{"$date":"2021-07-25T09:38:57.078+00:00"},"s":"I", "c":"COMMAND", "id":51803, "ctx":"conn541","msg":"Slow query","attr":{"type":"command","ns":"demo.hatchet","appName":"Keyhole Lib","command":{"aggregate":"hatchet","allowDiskUse":true,"pipeline":[{"$match":{"status":{"$in":["completed","split","splitting"]}}},{"$group":{"_id":{"replica_set":"$replica_set","namespace":"$query_filter.namespace"},"inserted":{"$sum":"$inserted"},"source_counts":{"$sum":"$source_counts"}}},{"$sort":{"status":1,"_id":-1}},{"$project":{"_id":0,"replica":"$_id.replica_set","ns":"$_id.namespace","inserted":1,"source_counts":1}}],"cursor":{},"lsid":{"id":{"$uuid":"86cf813b-463a-4e7b-b8f8-c587441a9575"}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1627205936,"i":4}},"signature":{"hash":{"$binary":{"base64":"Plz//gyzhsJMGIeEd6BdCIbgHSQ=","subType":"0"}},"keyId":6988792980442185732}},"$db":"_mongopush","$readPreference":{"mode":"primary"}},"planSummary":"IXSCAN { status: 1 }","keysExamined":218,"docsExamined":217,"hasSortStage":true,"cursorExhausted":true,"numYields":6,"nreturned":53,"queryHash":"6C0186CD","planCacheKey":"6EB1F22F","reslen":6117,"locks":{"ReplicationStateTransition":{"acquireCount":{"w":8}},"Global":{"acquireCount":{"r":8}},"Database":{"acquireCount":{"r":8}},"Collection":{"acquireCount":{"r":8}},"Mutex":{"acquireCount":{"r":2}}},"storage":{"data":{"bytesRead":4248700,"timeReadingMicros":527302}},"protocol":"op_msg","durationMillis":530}}`

// BenchmarkAnalyzeLog benchmarks the log parsing
func BenchmarkAnalyzeLog(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		AnalyzeLog(benchmarkLogStr)
	}
}

// BenchmarkBsonUnmarshal benchmarks BSON unmarshalling
func BenchmarkBsonUnmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		doc := Logv2Info{}
		bson.UnmarshalExtJSON([]byte(benchmarkLogStr), false, &doc)
	}
}

// BenchmarkBsonD2M benchmarks the optimized BsonD2M conversion
func BenchmarkBsonD2M(b *testing.B) {
	// Create a sample bson.D
	d := bson.D{
		{"first", "Ken"},
		{"last", "Chen"},
		{"nested", bson.D{{"a", 1}, {"b", 2}}},
		{"array", bson.A{1, 2, 3}},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BsonD2M(d)
	}
}

// BenchmarkBsonD2M_Old benchmarks the old Marshal/Unmarshal approach
func BenchmarkBsonD2M_Old(b *testing.B) {
	d := bson.D{
		{"first", "Ken"},
		{"last", "Chen"},
		{"nested", bson.D{{"a", 1}, {"b", 2}}},
		{"array", bson.A{1, 2, 3}},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Old approach using Marshal/Unmarshal
		var m bson.M
		data, _ := bson.Marshal(d)
		bson.Unmarshal(data, &m)
		_ = m
	}
}

// BenchmarkParseAttributes_Old benchmarks old attribute parsing (Marshal/Unmarshal)
func BenchmarkParseAttributes_Old(b *testing.B) {
	doc := Logv2Info{}
	bson.UnmarshalExtJSON([]byte(benchmarkLogStr), false, &doc)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Old approach
		data, _ := bson.Marshal(doc.Attr)
		bson.Unmarshal(data, &doc.Attributes)
	}
}

// BenchmarkAnalyzeSlowOp benchmarks slow operation analysis
func BenchmarkAnalyzeSlowOp(b *testing.B) {
	doc := Logv2Info{}
	bson.UnmarshalExtJSON([]byte(benchmarkLogStr), false, &doc)
	AddLegacyString(&doc)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeSlowOp(&doc)
	}
}

// BenchmarkParseAttributes benchmarks the optimized attribute parsing
func BenchmarkParseAttributes(b *testing.B) {
	doc := Logv2Info{}
	bson.UnmarshalExtJSON([]byte(benchmarkLogStr), false, &doc)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseAttributes(&doc)
	}
}

