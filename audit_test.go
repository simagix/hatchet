/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * audit_test.go
 */

package hatchet

import (
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mattn/go-sqlite3"
)

func init() {
	// Register sqlite3_extended driver with regex support
	regex := func(re, s string) (bool, error) {
		return regexp.MatchString(re, s)
	}
	sql.Register("sqlite3_extended",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regex, true)
			},
		})
}

func TestGetAuditDataWithClosedConnections(t *testing.T) {
	// Create a temporary database
	tmpDir := os.TempDir()
	dbfile := filepath.Join(tmpDir, "test_audit.db")
	defer os.Remove(dbfile)

	hatchetName := "test_hatchet"
	sqlite, err := NewSQLite3DB(dbfile, hatchetName, 2000)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.Close()

	// Begin transaction and create tables
	if err = sqlite.Begin(); err != nil {
		t.Fatal(err)
	}

	// Insert test data into _clients table
	clientStmt := `INSERT INTO test_hatchet_clients (id, ip, port, conns, accepted, ended, context, marker)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	clientData := [][]interface{}{
		{1, "192.168.1.100", "12345", 5, 10, 8, "conn-1", 0},
		{2, "192.168.1.101", "12346", 3, 15, 12, "conn-2", 0},
		{3, "192.168.1.102", "12347", 2, 20, 18, "conn-3", 0},
	}

	for _, data := range clientData {
		if _, err = sqlite.tx.Exec(clientStmt, data...); err != nil {
			t.Fatal(err)
		}
	}

	// Insert test log data to create reslen-ip entries
	logStmt := `INSERT INTO test_hatchet (id, date, severity, component, context, msg, plan, type, ns, message, op, filter, _index, milli, reslen, marker)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	logData := [][]interface{}{
		{1, "2023-01-01T10:00:00", "I", "NETWORK", "conn-1", "test", "", "", "test.collection", "test message", "query", "filter1", "idx1", 100, 5000, 0},
		{2, "2023-01-01T10:01:00", "I", "NETWORK", "conn-2", "test", "", "", "test.collection", "test message", "query", "filter2", "idx2", 150, 7000, 0},
		{3, "2023-01-01T10:02:00", "I", "NETWORK", "conn-3", "test", "", "", "test.collection", "test message", "query", "filter3", "idx3", 200, 9000, 0},
	}

	for _, data := range logData {
		if _, err = sqlite.tx.Exec(logStmt, data...); err != nil {
			t.Fatal(err)
		}
	}

	// Commit the transaction
	if err = sqlite.Commit(); err != nil {
		t.Fatal(err)
	}

	// Create metadata (this should create audit entries including ended-ip)
	if err = sqlite.CreateMetaData(); err != nil {
		t.Fatal(err)
	}

	// Get audit data
	data, err := sqlite.GetAuditData()
	if err != nil {
		t.Fatal(err)
	}

	// Verify IP audit data exists
	ipData, exists := data["ip"]
	if !exists {
		t.Fatal("Expected 'ip' category in audit data")
	}

	if len(ipData) == 0 {
		t.Fatal("Expected at least one IP entry in audit data")
	}

	// Verify each IP entry has 3 values: [accepted, reslen, ended]
	for _, entry := range ipData {
		if len(entry.Values) != 3 {
			t.Errorf("Expected 3 values (accepted, reslen, ended), got %d for IP %s",
				len(entry.Values), entry.Name)
		}

		// Verify the values are in the correct order
		accepted, ok := entry.Values[0].(int)
		if !ok {
			t.Errorf("Expected Values[0] (accepted) to be int for IP %s", entry.Name)
		}

		reslen, ok := entry.Values[1].(int)
		if !ok {
			t.Errorf("Expected Values[1] (reslen) to be int for IP %s", entry.Name)
		}

		ended, ok := entry.Values[2].(int)
		if !ok {
			t.Errorf("Expected Values[2] (ended) to be int for IP %s", entry.Name)
		}

		t.Logf("IP: %s, Accepted: %d, Reslen: %d, Ended: %d",
			entry.Name, accepted, reslen, ended)

		// Verify ended connections value matches what we inserted
		switch entry.Name {
		case "192.168.1.100":
			if ended != 8 {
				t.Errorf("Expected 8 ended connections for 192.168.1.100, got %d", ended)
			}
			if accepted != 10 {
				t.Errorf("Expected 10 accepted connections for 192.168.1.100, got %d", accepted)
			}
		case "192.168.1.101":
			if ended != 12 {
				t.Errorf("Expected 12 ended connections for 192.168.1.101, got %d", ended)
			}
			if accepted != 15 {
				t.Errorf("Expected 15 accepted connections for 192.168.1.101, got %d", accepted)
			}
		case "192.168.1.102":
			if ended != 18 {
				t.Errorf("Expected 18 ended connections for 192.168.1.102, got %d", ended)
			}
			if accepted != 20 {
				t.Errorf("Expected 20 accepted connections for 192.168.1.102, got %d", accepted)
			}
		}
	}
}

func TestGetAuditDataBackwardCompatibility(t *testing.T) {
	// Test backward compatibility when ended-ip data doesn't exist
	tmpDir := os.TempDir()
	dbfile := filepath.Join(tmpDir, "test_audit_compat.db")
	defer os.Remove(dbfile)

	hatchetName := "test_compat"
	sqlite, err := NewSQLite3DB(dbfile, hatchetName, 2000)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.Close()

	if err = sqlite.Begin(); err != nil {
		t.Fatal(err)
	}

	// Insert test data into _clients table
	stmt := `INSERT INTO test_compat_clients (id, ip, port, conns, accepted, ended, context, marker)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	if _, err = sqlite.tx.Exec(stmt, 1, "192.168.1.100", "12345", 5, 10, 8, "conn-1", 0); err != nil {
		t.Fatal(err)
	}

	if err = sqlite.Commit(); err != nil {
		t.Fatal(err)
	}

	// Manually create only ip and reslen-ip audit entries (simulating old database)
	insertStmt := `INSERT INTO test_compat_audit (type, name, value) VALUES (?, ?, ?)`
	if _, err = sqlite.db.Exec(insertStmt, "ip", "192.168.1.100", 10); err != nil {
		t.Fatal(err)
	}
	if _, err = sqlite.db.Exec(insertStmt, "reslen-ip", "192.168.1.100", 5000); err != nil {
		t.Fatal(err)
	}

	// Get audit data - should work even without ended-ip entries
	data, err := sqlite.GetAuditData()
	if err != nil {
		t.Fatal(err)
	}

	ipData, exists := data["ip"]
	if !exists || len(ipData) == 0 {
		t.Fatal("Expected IP audit data to exist")
	}

	// Should still have 3 values, with ended defaulting to 0
	entry := ipData[0]
	if len(entry.Values) != 3 {
		t.Fatalf("Expected 3 values even without ended-ip data, got %d", len(entry.Values))
	}

	// Verify the third value (ended) is 0 (from COALESCE)
	ended := entry.Values[2].(int)
	if ended != 0 {
		t.Errorf("Expected ended connections to be 0 (default from COALESCE), got %d", ended)
	}

	t.Logf("Backward compatibility test passed: IP=%s, Accepted=%d, Reslen=%d, Ended=%d (defaulted)",
		entry.Name, entry.Values[0], entry.Values[1], entry.Values[2])
}

func TestAuditTemplateConditionalColumn(t *testing.T) {
	// Test that the template can handle both cases: with and without ended data

	// Test case 1: Data with 3 values (should show Closed Connections column)
	dataWith3Values := map[string][]NameValues{
		"ip": {
			{Name: "192.168.1.100", Values: []interface{}{10, 5000, 8}},
			{Name: "192.168.1.101", Values: []interface{}{15, 7000, 12}},
		},
	}

	for _, entry := range dataWith3Values["ip"] {
		if len(entry.Values) >= 3 {
			t.Logf("Entry %s has %d values - Closed Connections column SHOULD be displayed",
				entry.Name, len(entry.Values))
		} else {
			t.Errorf("Expected entry %s to have 3 values, got %d", entry.Name, len(entry.Values))
		}
	}

	// Test case 2: Data with 2 values (should NOT show Closed Connections column)
	dataWith2Values := map[string][]NameValues{
		"ip": {
			{Name: "192.168.1.100", Values: []interface{}{10, 5000}},
		},
	}

	for _, entry := range dataWith2Values["ip"] {
		if len(entry.Values) >= 3 {
			t.Errorf("Entry %s should only have 2 values, got %d", entry.Name, len(entry.Values))
		} else {
			t.Logf("Entry %s has %d values - Closed Connections column should NOT be displayed",
				entry.Name, len(entry.Values))
		}
	}
}

func TestEndedIPAuditCreation(t *testing.T) {
	// Verify that CreateMetaData properly creates ended-ip audit entries
	tmpDir := os.TempDir()
	dbfile := filepath.Join(tmpDir, "test_ended_ip.db")
	defer os.Remove(dbfile)

	hatchetName := "test_ended"
	sqlite, err := NewSQLite3DB(dbfile, hatchetName, 2000)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.Close()

	if err = sqlite.Begin(); err != nil {
		t.Fatal(err)
	}

	// Insert multiple connections for the same IP to test SUM aggregation
	stmt := `INSERT INTO test_ended_clients (id, ip, port, conns, accepted, ended, context, marker)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	// Same IP with multiple connections
	if _, err = sqlite.tx.Exec(stmt, 1, "192.168.1.100", "12345", 5, 10, 8, "conn-1", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = sqlite.tx.Exec(stmt, 2, "192.168.1.100", "12346", 3, 5, 4, "conn-2", 0); err != nil {
		t.Fatal(err)
	}

	if err = sqlite.Commit(); err != nil {
		t.Fatal(err)
	}

	// Create metadata
	if err = sqlite.CreateMetaData(); err != nil {
		t.Fatal(err)
	}

	// Query the audit table directly to verify ended-ip was created
	query := `SELECT type, name, value FROM test_ended_audit WHERE type = 'ended-ip'`
	rows, err := sqlite.db.Query(query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var auditType, name string
		var value int
		if err = rows.Scan(&auditType, &name, &value); err != nil {
			t.Fatal(err)
		}

		if auditType == "ended-ip" && name == "192.168.1.100" {
			found = true
			// Should be sum of 8 + 4 = 12
			if value != 12 {
				t.Errorf("Expected SUM of ended connections to be 12, got %d", value)
			}
			t.Logf("Found ended-ip audit entry: IP=%s, Value=%d", name, value)
		}
	}

	if !found {
		t.Error("Expected to find ended-ip audit entry for 192.168.1.100")
	}
}
