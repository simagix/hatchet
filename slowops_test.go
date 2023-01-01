// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestAnalyzeSlowOp(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T09:38:57.078+00:00"},"s":"I",  "c":"COMMAND",  "id":51803,   "ctx":"conn541","msg":"Slow query","attr":{"type":"command","ns":"_mongopush.tasks","appName":"Keyhole Lib","command":{"aggregate":"tasks","allowDiskUse":true,"pipeline":[{"$match":{"status":{"$in":["completed","split","splitting"]}}},{"$group":{"_id":{"replica_set":"$replica_set","namespace":"$query_filter.namespace"},"inserted":{"$sum":"$inserted"},"source_counts":{"$sum":"$source_counts"}}},{"$sort":{"status":1,"_id":-1}},{"$project":{"_id":0,"replica":"$_id.replica_set","ns":"$_id.namespace","inserted":1,"source_counts":1}}],"cursor":{},"lsid":{"id":{"$uuid":"86cf813b-463a-4e7b-b8f8-c587441a9575"}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1627205936,"i":4}},"signature":{"hash":{"$binary":{"base64":"Plz//gyzhsJMGIeEd6BdCIbgHSQ=","subType":"0"}},"keyId":6988792980442185732}},"$db":"_mongopush","$readPreference":{"mode":"primary"}},"planSummary":"IXSCAN { status: 1 }","keysExamined":218,"docsExamined":217,"hasSortStage":true,"cursorExhausted":true,"numYields":6,"nreturned":53,"queryHash":"6C0186CD","planCacheKey":"6EB1F22F","reslen":6117,"locks":{"ReplicationStateTransition":{"acquireCount":{"w":8}},"Global":{"acquireCount":{"r":8}},"Database":{"acquireCount":{"r":8}},"Collection":{"acquireCount":{"r":8}},"Mutex":{"acquireCount":{"r":2}}},"storage":{"data":{"bytesRead":4248700,"timeReadingMicros":527302}},"protocol":"op_msg","durationMillis":530}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", doc.Timestamp.Format(time.RFC3339), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)

	stat, err := AnalyzeSlowOp(&doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(stat)
}
