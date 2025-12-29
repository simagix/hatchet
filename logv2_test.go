/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * logv2_test.go
 */

package hatchet

import (
	"os"
	"path/filepath"
	"testing"
)

// Note: sqlite3_extended driver is registered in audit_test.go init()

func TestAnalyze(t *testing.T) {
	filename := "testdata/mongod_ops.log.gz"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("skipping test: %s not found", filename)
	}
	logv2 := &Logv2{testing: true, url: SQLITE3_FILE}
	err := logv2.Analyze(filename, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzeLegacy(t *testing.T) {
	filename := "testdata/mongod_ops.log.gz"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("skipping test: %s not found", filename)
	}
	logv2 := &Logv2{testing: true, legacy: true}
	err := logv2.Analyze(filename, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsMongoDBLog(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "hatchet_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with valid MongoDB log
	validLog := `{"t":{"$date":"2021-07-25T09:56:00.691+00:00"},"s":"I","c":"WRITE","id":51803,"ctx":"conn1","msg":"Test message"}`
	validFile := filepath.Join(tmpDir, "valid.log")
	if err := os.WriteFile(validFile, []byte(validLog), 0644); err != nil {
		t.Fatal(err)
	}
	if !isMongoDBLog(validFile) {
		t.Error("expected valid MongoDB log to return true")
	}

	// Test with invalid file (not JSON)
	invalidFile := filepath.Join(tmpDir, "invalid.log")
	if err := os.WriteFile(invalidFile, []byte("not a json log"), 0644); err != nil {
		t.Fatal(err)
	}
	if isMongoDBLog(invalidFile) {
		t.Error("expected invalid file to return false")
	}

	// Test with JSON but not MongoDB log format
	jsonFile := filepath.Join(tmpDir, "json.log")
	if err := os.WriteFile(jsonFile, []byte(`{"name":"test","value":123}`), 0644); err != nil {
		t.Fatal(err)
	}
	if isMongoDBLog(jsonFile) {
		t.Error("expected non-MongoDB JSON to return false")
	}

	// Test with empty file
	emptyFile := filepath.Join(tmpDir, "empty.log")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if isMongoDBLog(emptyFile) {
		t.Error("expected empty file to return false")
	}

	// Test with non-existent file
	if isMongoDBLog(filepath.Join(tmpDir, "nonexistent.log")) {
		t.Error("expected non-existent file to return false")
	}
}

func TestAnalyzeDirectoryUniqueNames(t *testing.T) {
	// Create temp directory with multiple MongoDB log files
	tmpDir, err := os.MkdirTemp("", "hatchet_dir_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create 3 log files with same base name pattern
	logContent := `{"t":{"$date":"2021-07-25T09:56:00.691+00:00"},"s":"I","c":"WRITE","id":51803,"ctx":"conn1","msg":"Test message","attr":{"type":"insert","ns":"test.collection","durationMillis":100}}`
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(tmpDir, "mongodb.log")
		if i > 1 {
			filename = filepath.Join(tmpDir, "mongodb-"+string(rune('0'+i))+".log")
		}
		if err := os.WriteFile(filename, []byte(logContent+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test unique name generation for directory
	existingNames := []string{}
	processedNames := make([]string, 0)

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(tmpDir, entry.Name())
		if !isMongoDBLog(fullPath) {
			continue
		}
		allNames := append(existingNames, processedNames...)
		hatchetName := getUniqueHatchetName(fullPath, allNames)
		// Check for duplicates
		for _, name := range processedNames {
			if name == hatchetName {
				t.Errorf("duplicate hatchet name generated: %s", hatchetName)
			}
		}
		processedNames = append(processedNames, hatchetName)
		t.Logf("file: %s -> hatchet: %s", entry.Name(), hatchetName)
	}

	if len(processedNames) != 3 {
		t.Errorf("expected 3 unique hatchet names, got %d", len(processedNames))
	}

	// Verify all names are unique
	nameSet := make(map[string]bool)
	for _, name := range processedNames {
		if nameSet[name] {
			t.Errorf("duplicate name found: %s", name)
		}
		nameSet[name] = true
	}
}
