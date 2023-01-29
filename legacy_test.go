// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestAddLegacyStringAccess(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T10:08:39.336+00:00"},"s":"I",  "c":"ACCESS",   "id":20436,   "ctx":"conn1744","msg":"Checking authorization failed","attr":{"error":{"code":13,"codeName":"Unauthorized","errmsg":"not authorized on admin to execute command { setParameter: 1, ttlMonitorEnabled: true, lsid: { id: UUID(\"c6b35845-a51a-479e-8048-ca2b483e58b9\") }, $clusterTime: { clusterTime: Timestamp(1627207718, 8), signature: { hash: BinData(0, DF0FBE59451C5BDD4753901ACE41C5E072E59640), keyId: 6988792980442185732 } }, $db: \"admin\", $readPreference: { mode: \"primary\" } }"}}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", getDateTimeStr(doc.Timestamp), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringCommand(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T10:08:28.681+00:00"},"s":"I",  "c":"COMMAND",  "id":51803,   "ctx":"conn1725","msg":"Slow query","attr":{"type":"command","ns":"hatchet.favorites","appName":"hatchet","command":{"insert":"favorites","ordered":false,"writeConcern":{"w":"majority"},"lsid":{"id":{"$uuid":"80e1cb06-03b9-4b2f-83fc-61b920061bf0"}},"txnNumber":405,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1627207707,"i":860}},"signature":{"hash":{"$binary":{"base64":"KcbUE7Ii88QhQmdnM/X7BCdB7LM=","subType":"0"}},"keyId":6988792980442185732}},"$db":"hatchet"},"ninserted":6030,"keysInserted":12060,"numYields":0,"reslen":230,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":128}},"ReplicationStateTransition":{"acquireCount":{"w":129}},"Global":{"acquireCount":{"w":128}},"Database":{"acquireCount":{"w":128}},"Collection":{"acquireCount":{"w":128}},"Mutex":{"acquireCount":{"r":128}}},"flowControl":{"acquireCount":64,"timeAcquiringMicros":46},"writeConcern":{"w":"majority","wtimeout":0,"provenance":"clientSupplied"},"storage":{},"protocol":"op_msg","durationMillis":710}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", getDateTimeStr(doc.Timestamp), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringControl(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T10:08:38.558+00:00"},"s":"I",  "c":"COMMAND",  "id":51803,   "ctx":"conn1730","msg":"Slow query","attr":{"type":"command","ns":"hatchet.favorites","appName":"hatchet","command":{"insert":"favorites","ordered":false,"writeConcern":{"w":"majority"},"lsid":{"id":{"$uuid":"4c7dd848-651e-4ae3-a032-3a36ba4ad9fc"}},"txnNumber":297,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1627207717,"i":9267}},"signature":{"hash":{"$binary":{"base64":"M9YSQhYCG0p+gwxcKq2gn9LNOhc=","subType":"0"}},"keyId":6988792980442185732}},"$db":"hatchet"},"ninserted":3975,"keysInserted":7950,"numYields":0,"reslen":230,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":84}},"ReplicationStateTransition":{"acquireCount":{"w":85}},"Global":{"acquireCount":{"w":84}},"Database":{"acquireCount":{"w":84}},"Collection":{"acquireCount":{"w":84}},"Mutex":{"acquireCount":{"r":84}}},"flowControl":{"acquireCount":42,"timeAcquiringMicros":35},"writeConcern":{"w":"majority","wtimeout":0,"provenance":"clientSupplied"},"storage":{"data":{"bytesRead":25495,"timeReadingMicros":34}},"protocol":"op_msg","durationMillis":760}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", getDateTimeStr(doc.Timestamp), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringNetwork(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T10:10:16.116+00:00"},"s":"I",  "c":"NETWORK",  "id":22943,   "ctx":"listener","msg":"Connection accepted","attr":{"remote":"192.168.240.37:29402","connectionId":1907,"connectionCount":151}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", getDateTimeStr(doc.Timestamp), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringCommandCreateIndexes(t *testing.T) {
	str := `{"t":{"$date":"2020-08-21T20:39:17.211-04:00"},"s":"I",  "c":"COMMAND",  "id":51803,   "ctx":"conn49","msg":"Slow query","attr":{"type":"command","ns":"keyhole.numbers","command":{"createIndexes":"numbers","indexes":[{"key":{"a":1,"b":1},"name":"a_1_b_1"}],"lsid":{"id":{"$uuid":"e4565b38-b5bf-463d-bed8-2b478d75f683"}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1598056756,"i":1086}},"signature":{"hash":{"$binary":{"base64":"DrPHXR2GAe0G69zfjmQKcrhzJZA=","subType":"0"}},"keyId":6863601293718454275}},"$db":"keyhole","$readPreference":{"mode":"primaryPreferred"}},"numYields":0,"reslen":271,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":3}},"ReplicationStateTransition":{"acquireCount":{"w":6}},"Global":{"acquireCount":{"r":2,"w":4}},"Database":{"acquireCount":{"r":1,"w":3}},"Collection":{"acquireCount":{"w":1,"W":1}},"Mutex":{"acquireCount":{"r":4}}},"flowControl":{"acquireCount":3,"timeAcquiringMicros":2},"storage":{},"protocol":"op_msg","durationMillis":385}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", getDateTimeStr(doc.Timestamp), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestToLegacyString(t *testing.T) {
	str := `{"t":{"$date":"2020-08-21T20:39:17.211-04:00"},"s":"I",  "c":"COMMAND",  "id":51803, "ctx":"conn49","msg":"Slow query","attr":{"type":"command","ns":"keyhole.numbers","command":{"createIndexes":"numbers","indexes":[{"key":{"a":1,"b":1},"name":"a_1_b_1"}],"lsid":{"id":{"$uuid":"e4565b38-b5bf-463d-bed8-2b478d75f683"}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1598056756,"i":1086}},"signature":{"hash":{"$binary":{"base64":"DrPHXR2GAe0G69zfjmQKcrhzJZA=","subType":"0"}},"keyId":6863601293718454275}},"$db":"keyhole","$readPreference":{"mode":"primaryPreferred"}},"numYields":0,"reslen":271,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":3}},"ReplicationStateTransition":{"acquireCount":{"w":6}},"Global":{"acquireCount":{"r":2,"w":4}},"Database":{"acquireCount":{"r":1,"w":3}},"Collection":{"acquireCount":{"w":1,"W":1}},"Mutex":{"acquireCount":{"r":4}}},"flowControl":{"acquireCount":3,"timeAcquiringMicros":2},"storage":{},"protocol":"op_msg","durationMillis":385}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}
	t.Log(toLegacyString(doc.Attr.Map()["command"]))
}

func TestAddLegacyStringPipeline(t *testing.T) {
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
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", getDateTimeStr(doc.Timestamp), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringClientMeta(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T09:26:14.105+00:00"},"s":"I",  "c":"NETWORK",  "id":51800,   "ctx":"conn177","msg":"client metadata","attr":{"remote":"192.168.241.193:46318","client":"conn177","doc":{"driver":{"name":"NetworkInterfaceTL","version":"4.4.7"},"os":{"type":"Linux","name":"CentOS Linux release 7.9.2009 (Core)","architecture":"x86_64","version":"Kernel 3.10.0-1160.31.1.el7.x86_64"}}}}`
	t.Log(str)
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", getDateTimeStr(doc.Timestamp), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringLogicalSessionWrite(t *testing.T) {
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
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", getDateTimeStr(doc.Timestamp), doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}
