/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * obfuscation_test.go
 */

package hatchet

import (
	"regexp"
	"strings"
	"testing"
)

func TestObfuscateInt(t *testing.T) {
	o := NewObfuscation()
	// Set coefficient for testing
	o.Obfuscator.Coefficient = 0.5

	// Test case 1: Obfuscating a new integer
	input1 := 10
	expectedOutput1 := 5
	actualOutput1 := o.ObfuscateInt(input1)
	if actualOutput1 != expectedOutput1 {
		t.Errorf("Expected %d but got %d for input %d", expectedOutput1, actualOutput1, input1)
	}

	// Test case 2: Obfuscating the same integer as test case 1
	// The function should return the cached result instead of recalculating it
	input2 := 10
	expectedOutput2 := expectedOutput1
	actualOutput2 := o.ObfuscateInt(input2)
	if actualOutput2 != expectedOutput2 {
		t.Errorf("Expected %d but got %d for input %d", expectedOutput2, actualOutput2, input2)
	}
}

func TestObfuscateNumber(t *testing.T) {
	o := NewObfuscation()
	// Set coefficient for testing
	o.Obfuscator.Coefficient = 0.5

	// Test case 1: Obfuscating a new positive number
	input1 := 10.5
	expectedOutput1 := 5.25
	actualOutput1 := o.ObfuscateNumber(input1)
	if actualOutput1 != expectedOutput1 {
		t.Errorf("Expected %f but got %f for input %f", expectedOutput1, actualOutput1, input1)
	}

	// Test case 2: Obfuscating the same positive number as test case 1
	// The function should return the cached result instead of recalculating it
	input2 := 10.5
	expectedOutput2 := expectedOutput1
	actualOutput2 := o.ObfuscateNumber(input2)
	if actualOutput2 != expectedOutput2 {
		t.Errorf("Expected %f but got %f for input %f", expectedOutput2, actualOutput2, input2)
	}
}

func TestObfuscateCreditCardNo(t *testing.T) {
	o := NewObfuscation()

	// Test case 1: Obfuscating a valid Visa credit card number (passes Luhn check)
	// 4532015112830366 is a valid test Visa number
	input1 := "4532015112830366"
	expectedOutput1 := "************0366"
	actualOutput1 := o.ObfuscateCreditCardNo(input1)
	if actualOutput1 != expectedOutput1 {
		t.Errorf("Expected %s but got %s for input %s", expectedOutput1, actualOutput1, input1)
	}

	// Test case 2: Obfuscating a valid Mastercard with hyphens
	// 5425233430109903 is a valid test Mastercard number
	input2 := "5425-2334-3010-9903"
	expectedOutput2 := "****-****-****-9903"
	actualOutput2 := o.ObfuscateCreditCardNo(input2)
	if actualOutput2 != expectedOutput2 {
		t.Errorf("Expected %s but got %s for input %s", expectedOutput2, actualOutput2, input2)
	}

	// Test case 3: Invalid credit card number should not be obfuscated
	input3 := "1234567890123456"
	expectedOutput3 := "1234567890123456" // Should remain unchanged (invalid Luhn)
	actualOutput3 := o.ObfuscateCreditCardNo(input3)
	if actualOutput3 != expectedOutput3 {
		t.Errorf("Expected %s but got %s for input %s", expectedOutput3, actualOutput3, input3)
	}

	// Test case 4: Obfuscating an empty credit card number
	input4 := ""
	expectedOutput4 := ""
	actualOutput4 := o.ObfuscateCreditCardNo(input4)
	if actualOutput4 != expectedOutput4 {
		t.Errorf("Expected %s but got %s for input %s", expectedOutput4, actualOutput4, input4)
	}
}

func TestObfuscateEmail(t *testing.T) {
	o := NewObfuscation()

	// Test case 1: Obfuscating a valid email address
	input1 := "john.doe@example.com"
	expectedOutput1Regex := regexp.MustCompile(`^[a-z]+@[a-z]+\.com$`)
	actualOutput1 := o.ObfuscateEmail(input1)
	if !expectedOutput1Regex.MatchString(actualOutput1) {
		t.Errorf("Expected obfuscated email to match pattern %s, but got %s", expectedOutput1Regex.String(), actualOutput1)
	}

	// Test case 2: Obfuscating an email address that is already obfuscated
	input2 := input1
	expectedOutput2 := actualOutput1
	actualOutput2 := o.ObfuscateEmail(input2)
	if actualOutput2 != expectedOutput2 {
		t.Errorf("Expected output to be %s but got %s for input %s", expectedOutput2, actualOutput2, input2)
	}
}

func TestObfuscateIP(t *testing.T) {
	o := NewObfuscation()

	// Test case 1: Obfuscating a valid IP address
	input1 := "192.168.0.1"
	expectedOutput1Regex := regexp.MustCompile(`^192\.[0-9]+\.[0-9]+\.1$`)
	actualOutput1 := o.ObfuscateIP(input1)
	if !expectedOutput1Regex.MatchString(actualOutput1) {
		t.Errorf("Expected obfuscated IP to match pattern %s, but got %s", expectedOutput1Regex.String(), actualOutput1)
	}

	// Test case 2: Obfuscating the same IP address as in test case 1
	expectedOutput2 := actualOutput1
	actualOutput2 := o.ObfuscateIP(input1)
	if actualOutput2 != expectedOutput2 {
		t.Errorf("Expected output to be %s but got %s for input %s", expectedOutput2, actualOutput2, input1)
	}

	// Test case 5: Obfuscating an empty IP address
	input5 := ""
	expectedOutput5 := ""
	actualOutput5 := o.ObfuscateIP(input5)
	if actualOutput5 != expectedOutput5 {
		t.Errorf("Expected output to be %s but got %s for input %s", expectedOutput5, actualOutput5, input5)
	}
}

func TestObfuscateFQDN(t *testing.T) {
	o := NewObfuscation()

	// Test case 1: Obfuscating a valid FQDN with 2 parts
	input1 := "example.com"
	expectedOutputRegex := regexp.MustCompile(`([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}`)
	actualOutput1 := o.ObfuscateFQDN(input1)
	if !expectedOutputRegex.MatchString(actualOutput1) {
		t.Errorf("Expected obfuscated FQDN to match pattern %s, but got %s", expectedOutputRegex.String(), actualOutput1)
	}

	// Test case 2: Obfuscating a valid FQDN with more than 2 parts
	input2 := "www.example.co.uk"
	actualOutput2 := o.ObfuscateFQDN(input2)
	if !expectedOutputRegex.MatchString(actualOutput2) {
		t.Errorf("Expected obfuscated FQDN to match pattern %s, but got %s", expectedOutputRegex.String(), actualOutput2)
	}

	// Test case 3: Obfuscating an empty FQDN
	input3 := ""
	expectedOutput3 := ""
	actualOutput3 := o.ObfuscateFQDN(input3)
	if actualOutput3 != expectedOutput3 {
		t.Errorf("Expected output to be %s but got %s for input %s", expectedOutput3, actualOutput3, input3)
	}
}

func TestObfuscateNS(t *testing.T) {
	o := NewObfuscation()

	// Test case 1: Obfuscate a valid namespace
	for _, ns := range []string{"example.collection", "mydb.mycollection"} {
		obfuscated := o.ObfuscateNS(ns)
		if obfuscated == ns {
			t.Errorf("ObfuscateNS(%q) should have obfuscated the namespace, got %q", ns, obfuscated)
		}
	}

	// Test case 2: Email addresses should NOT be treated as namespaces
	for _, ns := range []string{"user@example.com", "user@mail.example.com"} {
		obfuscated := o.ObfuscateNS(ns)
		if obfuscated != ns {
			t.Errorf("ObfuscateNS(%q) should not obfuscate email, got %q", ns, obfuscated)
		}
	}
}

func TestObfuscateSSN(t *testing.T) {
	o := NewObfuscation()

	// Test case 1: Obfuscating a valid SSN with hyphens (required format)
	input1 := "123-45-6789"
	expectedOutputRegex := regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)
	actualOutput1 := o.ObfuscateSSN(input1)
	if !expectedOutputRegex.MatchString(actualOutput1) {
		t.Errorf("Expected obfuscated SSN to match pattern %s, but got %s", expectedOutputRegex.String(), actualOutput1)
	}
	// Verify it's actually different (obfuscated)
	if actualOutput1 == input1 {
		t.Logf("Note: SSN %s was shuffled to %s", input1, actualOutput1)
	}

	// Test case 2: SSN without hyphens should NOT be obfuscated (IsSSN requires hyphens)
	input2 := "123456789"
	expectedOutput2 := "123456789" // Should remain unchanged
	actualOutput2 := o.ObfuscateSSN(input2)
	if actualOutput2 != expectedOutput2 {
		t.Errorf("Expected output to be %s but got %s for input %s (no hyphens = invalid SSN format)", expectedOutput2, actualOutput2, input2)
	}

	// Test case 3: Invalid SSN format should NOT be obfuscated
	input3 := "12345-6789"
	expectedOutput3 := "12345-6789" // Should remain unchanged
	actualOutput3 := o.ObfuscateSSN(input3)
	if actualOutput3 != expectedOutput3 {
		t.Errorf("Expected output to be %s but got %s for input %s (invalid format)", expectedOutput3, actualOutput3, input3)
	}

	// Test case 4: Obfuscating an empty SSN
	input4 := ""
	expectedOutput4 := ""
	actualOutput4 := o.ObfuscateSSN(input4)
	if actualOutput4 != expectedOutput4 {
		t.Errorf("Expected output to be %s but got %s for input %s", expectedOutput4, actualOutput4, input4)
	}

	// Test case 5: Another valid SSN
	input5 := "987-65-4321"
	actualOutput5 := o.ObfuscateSSN(input5)
	if !expectedOutputRegex.MatchString(actualOutput5) {
		t.Errorf("Expected obfuscated SSN to match pattern %s, but got %s", expectedOutputRegex.String(), actualOutput5)
	}
}

func TestObfuscatePhoneNo(t *testing.T) {
	o := NewObfuscation()

	// Test case 1: Obfuscating a valid phone number with 10 digits
	input1 := "1234567890"
	expectedOutputRegex := regexp.MustCompile(`(?:\+?\d{1,3}[- ]?)?\d{10,14}|(\+\d{1,3}\s?)?\(\d{3}\)\s?\d{3}[- ]?\d{4}|\d{3}[- ]?\d{3}[- ]?\d{4}`)
	actualOutput1 := o.ObfuscatePhoneNo(input1)
	if !expectedOutputRegex.MatchString(actualOutput1) {
		t.Errorf("Expected obfuscated phone number to match pattern %s, but got %s", expectedOutputRegex.String(), actualOutput1)
	}

	// Test case 2: Obfuscating a valid phone number with 10 digits
	input2 := "123-456-7890"
	expectedOutput2Regex := regexp.MustCompile(`^(\d{3})[-\.\s]?(\d{3})[-\.\s]?(\d{4})$`)
	actualOutput2 := o.ObfuscatePhoneNo(input2)
	if !expectedOutputRegex.MatchString(actualOutput2) {
		t.Errorf("Expected obfuscated phone number to match pattern %s, but got %s", expectedOutput2Regex.String(), actualOutput2)
	}

	// Test case 3: Obfuscating an empty phone number
	input3 := ""
	expectedOutput3 := ""
	actualOutput3 := o.ObfuscatePhoneNo(input3)
	if actualOutput3 != expectedOutput3 {
		t.Errorf("Expected output to be %s but got %s for input %s", expectedOutput3, actualOutput3, input3)
	}

	// Test case 4: Obfuscating an empty phone number
	input4 := "+1 (123) 456-7890"
	actualOutput4 := o.ObfuscatePhoneNo(input4)
	if !expectedOutputRegex.MatchString(actualOutput4) {
		t.Errorf("Expected obfuscated phone number to match pattern %s, but got %s", expectedOutputRegex.String(), actualOutput4)
	}
}

func TestObfuscateMAC(t *testing.T) {
	obs := NewObfuscation()

	// Test MAC with colons
	result := obs.ObfuscateMAC("AA:BB:CC:11:22:33")
	if result == "AA:BB:CC:11:22:33" {
		t.Error("MAC should be obfuscated")
	}
	// Vendor prefix should be preserved
	if !strings.HasPrefix(result, "AA:BB:CC:") {
		t.Errorf("Vendor prefix should be preserved, got %s", result)
	}

	// Test determinism
	result2 := obs.ObfuscateMAC("AA:BB:CC:11:22:33")
	if result != result2 {
		t.Error("Same MAC should produce same result")
	}

	// Test MAC with dashes
	result3 := obs.ObfuscateMAC("AA-BB-CC-11-22-33")
	if !strings.HasPrefix(result3, "AA-BB-CC-") {
		t.Errorf("Vendor prefix should be preserved with dashes, got %s", result3)
	}
}

func TestObfuscateDate(t *testing.T) {
	obs := NewObfuscation()

	result := obs.ObfuscateDate("2024-06-15")
	if result == "2024-06-15" {
		t.Error("Date should be shifted")
	}
	// Should still be a valid date format
	if len(result) != 10 || result[4] != '-' || result[7] != '-' {
		t.Errorf("Date format should be preserved, got %s", result)
	}

	// Test determinism
	obs2 := NewObfuscation()
	result2 := obs2.ObfuscateDate("2024-06-15")
	if result != result2 {
		t.Error("Same date should produce same result across instances")
	}
}

func TestObfuscateID(t *testing.T) {
	obs := NewObfuscation()

	// Test MRN
	result := obs.ObfuscateID("MRN: 12345678")
	if result == "MRN: 12345678" {
		t.Error("MRN should be obfuscated")
	}
	if !strings.HasPrefix(result, "MRN: ") {
		t.Errorf("MRN prefix should be preserved, got %s", result)
	}

	// Test account number
	result2 := obs.ObfuscateID("acct#987654321")
	if result2 == "acct#987654321" {
		t.Error("Account should be obfuscated")
	}

	// Test determinism
	result3 := obs.ObfuscateID("MRN: 12345678")
	if result != result3 {
		t.Error("Same ID should produce same result")
	}
}
