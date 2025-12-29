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
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	maxUploadSize   = 200 << 20 // 200 MB max file size
	maxUploadSizeMB = 200
)

// UploadHandler handles file uploads for processing
func UploadHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": "Method not allowed"})
		return
	}

	// Parse multipart form - 32MB max in memory, rest goes to temp files
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": fmt.Sprintf("Failed to parse form: %v", err)})
		return
	}

	file, header, err := r.FormFile("logfile")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": fmt.Sprintf("Failed to get file: %v", err)})
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > maxUploadSize {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("File too large (%d MB). Maximum size: %d MB", header.Size>>20, maxUploadSizeMB),
		})
		return
	}

	// Get optional hatchet name from form, default to filename
	hatchetName := r.FormValue("name")
	if hatchetName == "" {
		hatchetName = header.Filename
		// Remove common extensions
		hatchetName = strings.TrimSuffix(hatchetName, ".gz")
		hatchetName = strings.TrimSuffix(hatchetName, ".log")
	}
	// Get unique name (adds _2, _3 suffix if name exists, like CLI does)
	existingNames, _ := GetExistingHatchetNames()
	hatchetName = getUniqueHatchetName(hatchetName, existingNames)

	// Create temp file to store the upload
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "hatchet-upload-*")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": fmt.Sprintf("Failed to create temp file: %v", err)})
		return
	}
	tempPath := tempFile.Name()

	// Stream upload to temp file (memory efficient)
	written, err := io.Copy(tempFile, file)
	tempFile.Close()
	if err != nil {
		os.Remove(tempPath)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": fmt.Sprintf("Failed to save file: %v", err)})
		return
	}

	log.Printf("Uploaded %s (%d bytes) -> %s", header.Filename, written, tempPath)

	// Validate file is a MongoDB log by checking content
	if !isMongoDBLog(tempPath) {
		os.Remove(tempPath)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Not a valid MongoDB log file. File must contain MongoDB logv2 JSON format.",
		})
		return
	}

	// Create a new Logv2 instance for this upload to support concurrent uploads
	// Copy config from the singleton but use separate state
	baseLogv2 := GetLogv2()
	uploadLogv2 := &Logv2{
		url:         baseLogv2.url,
		hatchetName: hatchetName,
		version:     baseLogv2.version,
		cacheSize:   baseLogv2.cacheSize,
		from:        time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), // Include all past logs
		to:          time.Now().Add(24 * time.Hour),              // Include logs up to tomorrow
	}

	go func(name string, logv2 *Logv2) {
		defer os.Remove(tempPath) // Clean up temp file when done

		if err := logv2.Analyze(tempPath, 1); err != nil {
			log.Printf("Error processing upload %s: %v", header.Filename, err)
			return
		}
		logv2.PrintSummary()
		log.Printf("Finished processing upload: %s -> %s", header.Filename, name)
	}(hatchetName, uploadLogv2)

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
