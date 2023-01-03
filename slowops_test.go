// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"testing"

	"github.com/simagix/gox"
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

	stat, err := AnalyzeSlowOp(&doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(gox.Stringify(stat))
}

func TestAnalyzeSlowOpUpdate(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T09:56:00.691+00:00"},"s":"I",  "c":"WRITE",    "id":51803,   "ctx":"LogicalSessionCacheRefresh","msg":"Slow query","attr":{"type":"update","ns":"config.system.sessions","command":{"q":{"_id":{"id":{"$uuid":"6712143d-c644-4e18-a627-555ac42f35e5"},"uid":{"$binary":{"base64":"FS5Vi3aeniqLFs3ALoTFS1pJY/Sz3Ngs1h+xZYOrI8Y=","subType":"0"}}}},"u":[{"$set":{"lastUse":"$$NOW"}}],"multi":false,"upsert":true},"planSummary":"IDHACK","keysExamined":0,"docsExamined":0,"nMatched":0,"nModified":0,"upsert":true,"keysInserted":2,"numYields":0,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":1}},"ReplicationStateTransition":{"acquireCount":{"w":1}},"Global":{"acquireCount":{"w":1}},"Database":{"acquireCount":{"w":1}},"Collection":{"acquireCount":{"w":1}},"Mutex":{"acquireCount":{"r":1}}},"flowControl":{"acquireCount":1,"timeAcquiringMicros":1},"storage":{"data":{"bytesRead":23074,"timeReadingMicros":105638}},"durationMillis":105}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}

	stat, err := AnalyzeSlowOp(&doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(gox.Stringify(stat))
}

func TestAnalyzeSlowOpFind(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T09:26:14.284+00:00"},"s":"I",  "c":"COMMAND",  "id":51803,   "ctx":"conn177","msg":"Slow query","attr":{"type":"command","ns":"local.oplog.rs","command":{"find":"oplog.rs","limit":1,"sort":{"$natural":1},"projection":{"ts":1,"t":1},"readConcern":{},"$readPreference":{"mode":"secondaryPreferred"},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1627205174,"i":1}},"signature":{"hash":{"$binary":{"base64":"XNbnSLT6HhfEZh3zaPR9bJud+0E=","subType":"0"}},"keyId":6988792980442185732}},"$db":"local"},"planSummary":"COLLSCAN","keysExamined":0,"docsExamined":1,"cursorExhausted":true,"numYields":0,"nreturned":1,"reslen":259,"locks":{"ReplicationStateTransition":{"acquireCount":{"w":1},"acquireWaitCount":{"w":1},"timeAcquiringMicros":{"w":113092}},"Global":{"acquireCount":{"r":1}},"Database":{"acquireCount":{"r":1}},"Mutex":{"acquireCount":{"r":1}},"oplog":{"acquireCount":{"r":1}}},"readConcern":{"provenance":"clientSupplied"},"storage":{},"protocol":"op_msg","durationMillis":114}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}

	stat, err := AnalyzeSlowOp(&doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(gox.Stringify(stat))
}
