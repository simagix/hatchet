/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * drivers_test.go
 */

package hatchet

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestCSharpDriverVersion(t *testing.T) {
}

func TestGoDriverVersion(t *testing.T) {
	str := `{"t":{"$date":"2023-03-25T16:12:23.482+00:00"},"s":"I",  "c":"NETWORK",  "id":51800,   "ctx":"conn1865","msg":"client metadata","attr":{"remote":"192.168.0.123:39706","client":"conn1865","doc":{"driver":{"name":"mongo-go-driver","version":"v1.5.2"},"os":{"type":"linux","architecture":"amd64"},"platform":"go1.15.9","application":{"name":"MongoDB CPS Module v11.1.2.6974 (git: ec985434025211bab94ccfa89cd4318f36e301c3)"}}}}`
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}
	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	expected := "mongo-go-driver"
	if doc.Client.Driver != expected {
		t.Fatal("expected", expected, "but got", doc.Client.Driver)
	}
	if err = CheckDriverCompatibility("v4.4", expected, doc.Client.Version); err != nil {
		t.Fatal(err)
	}
}

func TestJavaDriverVersion(t *testing.T) {
}

func TestNodeJSDriverVersion(t *testing.T) {
	str := `{"t":{"$date":"2023-03-25T16:12:23.482+00:00"},"s":"I", "c":"NETWORK", "id":51800, "ctx":"conn67009","msg":"client metadata","attr":{"remote":"192.168.0.123:47790","client":"conn67009","doc":{"driver":{"name":"nodejs","version":"4.1.0"},"os":{"type":"Windows_NT","name":"win32","architecture":"x64","version":"10.0.19044"},"platform":"Node.js v12.4.0, LE (unified)|Node.js v12.4.0, LE (unified)","application":{"name":"MongoDB Compass"}}}}`
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}
	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	expected := "nodejs"
	if doc.Client.Driver != expected {
		t.Fatal("expected", expected, "but got", doc.Client.Driver)
	}
	if err = CheckDriverCompatibility("v4.4", expected, doc.Client.Version); err != nil {
		t.Fatal(err)
	}
}

func TestPythonDriverVersion(t *testing.T) {
	str := `{"t":{"$date":"2023-03-25T16:12:23.482+00:00"},"s":"I", "c":"NETWORK", "id":51800, "ctx":"conn16","msg":"client metadata","attr":{"remote":"127.0.0.1:62373","client":"conn16","doc":{"driver":{"name":"PyMongo","version":"3.11.0"},"os":{"type":"Darwin","name":"Darwin","architecture":"x86_64","version":"12.6.3"},"platform":"CPython 3.8.16.final.0","application":{"name":"mlaunch v1.6.4"}}}}`
	doc := Logv2Info{}
	err := bson.UnmarshalExtJSON([]byte(str), false, &doc)
	if err != nil {
		t.Fatalf("bson unmarshal error %v", err)
	}
	if err = AddLegacyString(&doc); err != nil {
		t.Fatalf("logv2 marshal error %v", err)
	}
	expected := "PyMongo"
	if doc.Client.Driver != expected {
		t.Fatal("expected", expected, "but got", doc.Client.Driver)
	}
	if err = CheckDriverCompatibility("v4.4", expected, doc.Client.Version); err != nil {
		t.Fatal(err)
	}
}

func TestRubyDriverVersion(t *testing.T) {
}
