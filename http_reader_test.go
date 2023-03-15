/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * http_reader_test.go
 */

package hatchet

import (
	"io"
	"testing"
)

func TestGetHTTPContent(t *testing.T) {
	reader, err := GetHTTPContent("http://localhost:9090/testdata/my-file.txt", "", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t.Log(string(content))
}
