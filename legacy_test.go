// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestAddLegacyStringAccess(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T10:08:39.336+00:00"},"s":"I",  "c":"ACCESS",   "id":20436,   "ctx":"conn1744","msg":"Checking authorization failed","attr":{"error":{"code":13,"codeName":"Unauthorized","errmsg":"not authorized on admin to execute command { setParameter: 1, ttlMonitorEnabled: true, lsid: { id: UUID(\"c6b35845-a51a-479e-8048-ca2b483e58b9\") }, $clusterTime: { clusterTime: Timestamp(1627207718, 8), signature: { hash: BinData(0, DF0FBE59451C5BDD4753901ACE41C5E072E59640), keyId: 6988792980442185732 } }, $db: \"admin\", $readPreference: { mode: \"primary\" } }"}}}`
	t.Log(str)
	doc := Logv2Info{}
	err := json.Unmarshal([]byte(str), &doc)
	if err != nil {
		t.Fatalf("json unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", doc.Timestamp["$date"], doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringCommand(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T10:08:28.681+00:00"},"s":"I",  "c":"COMMAND",  "id":51803,   "ctx":"conn1725","msg":"Slow query","attr":{"type":"command","ns":"hatchet.favorites","appName":"hatchet","command":{"insert":"favorites","ordered":false,"writeConcern":{"w":"majority"},"lsid":{"id":{"$uuid":"80e1cb06-03b9-4b2f-83fc-61b920061bf0"}},"txnNumber":405,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1627207707,"i":860}},"signature":{"hash":{"$binary":{"base64":"KcbUE7Ii88QhQmdnM/X7BCdB7LM=","subType":"0"}},"keyId":6988792980442185732}},"$db":"hatchet"},"ninserted":6030,"keysInserted":12060,"numYields":0,"reslen":230,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":128}},"ReplicationStateTransition":{"acquireCount":{"w":129}},"Global":{"acquireCount":{"w":128}},"Database":{"acquireCount":{"w":128}},"Collection":{"acquireCount":{"w":128}},"Mutex":{"acquireCount":{"r":128}}},"flowControl":{"acquireCount":64,"timeAcquiringMicros":46},"writeConcern":{"w":"majority","wtimeout":0,"provenance":"clientSupplied"},"storage":{},"protocol":"op_msg","durationMillis":710}}`
	t.Log(str)
	doc := Logv2Info{}
	err := json.Unmarshal([]byte(str), &doc)
	if err != nil {
		t.Fatalf("json unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", doc.Timestamp["$date"], doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringControl(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T10:08:38.558+00:00"},"s":"I",  "c":"COMMAND",  "id":51803,   "ctx":"conn1730","msg":"Slow query","attr":{"type":"command","ns":"hatchet.favorites","appName":"hatchet","command":{"insert":"favorites","ordered":false,"writeConcern":{"w":"majority"},"lsid":{"id":{"$uuid":"4c7dd848-651e-4ae3-a032-3a36ba4ad9fc"}},"txnNumber":297,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1627207717,"i":9267}},"signature":{"hash":{"$binary":{"base64":"M9YSQhYCG0p+gwxcKq2gn9LNOhc=","subType":"0"}},"keyId":6988792980442185732}},"$db":"hatchet"},"ninserted":3975,"keysInserted":7950,"numYields":0,"reslen":230,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":84}},"ReplicationStateTransition":{"acquireCount":{"w":85}},"Global":{"acquireCount":{"w":84}},"Database":{"acquireCount":{"w":84}},"Collection":{"acquireCount":{"w":84}},"Mutex":{"acquireCount":{"r":84}}},"flowControl":{"acquireCount":42,"timeAcquiringMicros":35},"writeConcern":{"w":"majority","wtimeout":0,"provenance":"clientSupplied"},"storage":{"data":{"bytesRead":25495,"timeReadingMicros":34}},"protocol":"op_msg","durationMillis":760}}`
	t.Log(str)
	doc := Logv2Info{}
	err := json.Unmarshal([]byte(str), &doc)
	if err != nil {
		t.Fatalf("json unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", doc.Timestamp["$date"], doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}

func TestAddLegacyStringNetwork(t *testing.T) {
	str := `{"t":{"$date":"2021-07-25T10:10:16.116+00:00"},"s":"I",  "c":"NETWORK",  "id":22943,   "ctx":"listener","msg":"Connection accepted","attr":{"remote":"192.168.240.37:29402","connectionId":1907,"connectionCount":151}}`
	t.Log(str)
	doc := Logv2Info{}
	err := json.Unmarshal([]byte(str), &doc)
	if err != nil {
		t.Fatalf("json unmarshal error %v", err)
	}

	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	logstr := fmt.Sprintf("%v %-2s %-8s [%v] %v", doc.Timestamp["$date"], doc.Severity, doc.Component, doc.Context, doc.Message)
	t.Log(logstr)
}
