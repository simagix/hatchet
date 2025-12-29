/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * upload_handler.go
 */

package hatchet

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// UploadHandler handles file uploads for processing
func UploadHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form - 32MB max in memory, rest goes to temp files
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("logfile")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get optional hatchet name from form, default to filename
	hatchetName := r.FormValue("name")
	if hatchetName == "" {
		hatchetName = header.Filename
		// Remove common extensions
		hatchetName = strings.TrimSuffix(hatchetName, ".gz")
		hatchetName = strings.TrimSuffix(hatchetName, ".zst")
		hatchetName = strings.TrimSuffix(hatchetName, ".log")
	}
	hatchetName = getHatchetName(hatchetName)

	// Check if hatchet already exists
	existingNames, _ := GetExistingHatchetNames()
	for _, name := range existingNames {
		if name == hatchetName {
			http.Error(w, fmt.Sprintf("Hatchet '%s' already exists", hatchetName), http.StatusConflict)
			return
		}
	}

	// Create temp file to store the upload
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "hatchet-upload-*")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create temp file: %v", err), http.StatusInternalServerError)
		return
	}
	tempPath := tempFile.Name()

	// Stream upload to temp file (memory efficient)
	written, err := io.Copy(tempFile, file)
	tempFile.Close()
	if err != nil {
		os.Remove(tempPath)
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Uploaded %s (%d bytes) -> %s", header.Filename, written, tempPath)

	// Process the file using existing Logv2 analyzer
	// This runs in parallel automatically
	logv2 := GetLogv2()
	logv2.hatchetName = hatchetName

	go func(name string) {
		defer os.Remove(tempPath)                 // Clean up temp file when done
		defer func() { logv2.hatchetName = "" }() // Reset for next upload

		if err := logv2.Analyze(tempPath, 1); err != nil {
			log.Printf("Error processing upload %s: %v", header.Filename, err)
			return
		}
		logv2.PrintSummary()
		log.Printf("Finished processing upload: %s -> %s", header.Filename, name)
	}(hatchetName)

	// Return immediately with status
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "processing",
		"name":    hatchetName,
		"size":    written,
		"message": fmt.Sprintf("File '%s' uploaded and processing started", header.Filename),
	})
}

// UploadStatusHandler checks the status of a processing job
func UploadStatusHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hatchetName := params.ByName("name")

	existingNames, _ := GetExistingHatchetNames()
	for _, name := range existingNames {
		if name == hatchetName {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "complete",
				"name":   hatchetName,
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "processing",
		"name":   hatchetName,
	})
}

// getUploadDir ensures upload directory exists and returns path
func getUploadDir() string {
	dir := filepath.Join(".", "data", "uploads")
	os.MkdirAll(dir, 0755)
	return dir
}
