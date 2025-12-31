/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * report.go
 */

package hatchet

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const HTML_DIR = "./html"

// GenerateReports generates HTML reports for a hatchet
func GenerateReports(hatchetName string) error {
	// Create output directory
	if err := os.MkdirAll(HTML_DIR, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	dbase, err := GetDatabase(hatchetName)
	if err != nil {
		return fmt.Errorf("failed to get database: %v", err)
	}
	defer dbase.Close()

	info := dbase.GetHatchetInfo()
	summary := GetHatchetSummary(info)
	version := GetLogv2().version

	// Generate audit report
	if err := generateAuditReport(dbase, hatchetName, info, summary, version); err != nil {
		return fmt.Errorf("failed to generate audit report: %v", err)
	}

	// Generate stats report (slow ops patterns)
	if err := generateStatsReport(dbase, hatchetName, info, summary, version); err != nil {
		return fmt.Errorf("failed to generate stats report: %v", err)
	}

	// Generate top N slowest operations report
	if err := generateTopNReport(dbase, hatchetName, info, summary, version); err != nil {
		return fmt.Errorf("failed to generate top N report: %v", err)
	}

	log.Printf("HTML reports generated in %s/", HTML_DIR)
	return nil
}

func generateAuditReport(dbase Database, hatchetName string, info HatchetInfo, summary string, version string) error {
	data, err := dbase.GetAuditData()
	if err != nil {
		return err
	}

	templ, err := GetAuditTablesTemplate("true") // download mode for standalone HTML
	if err != nil {
		return err
	}

	doc := map[string]interface{}{
		"Hatchet": hatchetName,
		"Info":    info,
		"Summary": summary,
		"Data":    data,
		"Version": version,
	}

	var buf bytes.Buffer
	if err = templ.Execute(&buf, doc); err != nil {
		return err
	}

	filename := filepath.Join(HTML_DIR, fmt.Sprintf("%s_audit.html", hatchetName))
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return err
	}
	log.Printf("Generated %s", filename)
	return nil
}

func generateStatsReport(dbase Database, hatchetName string, info HatchetInfo, summary string, version string) error {
	// Get slow ops sorted by avg_ms DESC
	ops, err := dbase.GetSlowOps("avg_ms", "DESC", false)
	if err != nil {
		return err
	}

	templ, err := GetStatsTableTemplate(false, "avg_ms", "true") // download mode
	if err != nil {
		return err
	}

	doc := map[string]interface{}{
		"Hatchet": hatchetName,
		"Merge":   info.Merge,
		"Ops":     ops,
		"Summary": summary,
		"Version": version,
	}

	var buf bytes.Buffer
	if err = templ.Execute(&buf, doc); err != nil {
		return err
	}

	filename := filepath.Join(HTML_DIR, fmt.Sprintf("%s_stats.html", hatchetName))
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return err
	}
	log.Printf("Generated %s", filename)
	return nil
}

func generateTopNReport(dbase Database, hatchetName string, info HatchetInfo, summary string, version string) error {
	logstrs, err := dbase.GetSlowestLogs(TOP_N)
	if err != nil {
		return err
	}

	templ, err := GetLogTableTemplate("slowops", "true") // download mode
	if err != nil {
		return err
	}

	doc := map[string]interface{}{
		"Hatchet": hatchetName,
		"Merge":   info.Merge,
		"Logs":    logstrs,
		"Summary": summary,
		"Version": version,
	}

	var buf bytes.Buffer
	if err = templ.Execute(&buf, doc); err != nil {
		return err
	}

	filename := filepath.Join(HTML_DIR, fmt.Sprintf("%s_topn.html", hatchetName))
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return err
	}
	log.Printf("Generated %s", filename)
	return nil
}
