// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestToInt(t *testing.T) {
	input := "23"
	value := ToInt(input)
	if value != 23 {
		t.Fatal("expected", 23, "but got", value)
	}

	str := ""
	value = ToInt(str)
	if value != 0 {
		t.Fatal("expected", 0, "but got", value)
	}
}

func TestReplaceSpecialChars(t *testing.T) {
	value := "a_b_c_d_e_name"

	filename := "a-b.c d:e,name"
	fname := replaceSpecialChars(filename)
	if value != fname {
		t.Fatal("expected", value, "but got", fname)
	}
}

func TestGetHatchetName(t *testing.T) {
	filename := "mongod.log"
	length := len(filename) - len(".log") + TAIL_SIZE
	hatchetName := getHatchetName(filename)
	if len(hatchetName) != length {
		t.Fatal("expected", length, "but got", len(hatchetName))
	}

	filename = "mongod.log.gz"
	length = len(filename) - len(".log.gz") + TAIL_SIZE
	hatchetName = getHatchetName(filename)
	if len(hatchetName) != length {
		t.Fatal("expected", length, "but got", len(hatchetName))
	}

	filename = "mongod"
	length = len(filename) + TAIL_SIZE
	hatchetName = getHatchetName(filename)
	if len(hatchetName) != length {
		t.Fatal("expected", length, "but got", len(hatchetName))
	}

	filename = "filesys-shard-00-01.abcde.mongodb.net_2021-07-24T10_12_58_2021-07-25T10_12_58_mongodb.log.gz"
	length = len(filename) + TAIL_SIZE
	hatchetName = getHatchetName(filename)
	modified := "filesys_shard_00_01_abcde_mongodb_net_2021_07_24T10_12_58_2021_07_25T10_12_58_mongodb"
	if !strings.HasPrefix(hatchetName, modified) {
		t.Fatal(modified+"_*", length, "but got", hatchetName)
	}

	filename = "testdata/demo_errmsg.log.gz"
	fname := filepath.Base((filename))
	t.Log(fname)
	hatchetName = getHatchetName(filename)
	t.Log(hatchetName)
	if len(hatchetName) != len(fname) {
		t.Fatal("expected", len(fname), "but got", len(hatchetName))
	}

	filename = "testdata/0_errmsg.log.gz"
	fname = filepath.Base((filename))
	t.Log(fname)
	hatchetName = getHatchetName(filename)
	t.Log(hatchetName)
	if len(hatchetName) != len(fname)+1 { // added _
		t.Fatal("expected", len(fname), "but got", len(hatchetName))
	}
}

func TestGetDateSubString(t *testing.T) {
	value := GetSQLDateSubString("2023-01-01T12:11:02Z", "2023-01-01T14:35:20Z")
	if value != "SUBSTR(date, 1, 15)||'9:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 15)||'9:59'", "but got", value)
	}
	t.Log(value)

	value = GetSQLDateSubString("2023-01-01T12:11:02Z", "2023-01-02T12:34:20Z")
	if value != "SUBSTR(date, 1, 13)||':59:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 13)||':59:59'", "but got", value)
	}
	t.Log(value)

	value = GetSQLDateSubString("2023-01-01T12:11:02Z", "2023-02-10T12:34:20Z")
	if value != "SUBSTR(date, 1, 10)||'T23:59:59'" {
		t.Fatal("expected", "SUBSTR(date, 1, 13)||':59:59'", "but got", value)
	}
	t.Log(value)
}

func TestGetOffsetLimit(t *testing.T) {
	limit := "100"
	o, l := GetOffsetLimit(limit)
	if o != 0 || l != 100 {
		t.Fatal("expected", 0, 100, "but got", o, l)
	}

	limit = "100,100"
	o, l = GetOffsetLimit(limit)
	if o != 100 || l != 100 {
		t.Fatal("expected", 100, 100, "but got", o, l)
	}
}

func TestContainsCreditCardNo(t *testing.T) {
	validCases := []struct {
		input    string
		expected bool
	}{
		{"4111111111111111", true}, // Visa
		{"4012888888881881", true}, // Visa
		{"4012 8888 8888 1881", true},
		{"4222222222222", true},       // Visa
		{"4917610000000000003", true}, // Visa
		{"5105105105105100", true},    // Mastercard
		{"5555555555554444", true},    // Mastercard
		{"6011111111111117", true},    // Discover
		{"6011000990139424", true},    // Discover
		{"371449635398431", true},     // American Express
		{"378282246310005", true},     // American Express
		{"30569309025904", true},      // Diners Club
		{"38520000023237", true},      // Diners Club
		{"3530111333300000", true},    // JCB
	}

	invalidCases := []struct {
		input    string
		expected bool
	}{
		{"1234567890123456", false},
		{"1234 5678 9012 3456", false},
		{"1234567", false},
		{"4111-1111-1111-1112", false},
		{"4222222222222000", false},
		{"4917610000000000004", false},
		{"4917610000000000003000", false},
		{"5105105105105101", false},
		{"30569309025905", false},
		{"3530111333300001", false},
	}

	// Iterate over the valid test cases
	for _, tc := range validCases {
		result := ContainsCreditCardNo(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsCreditCardNo(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}

	// Iterate over the invalid test cases
	for _, tc := range invalidCases {
		result := ContainsCreditCardNo(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsCreditCardNo(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}
}

func TestContainsEmailAddress(t *testing.T) {
	validCases := []struct {
		input    string
		expected bool
	}{
		{"test@example.com", true},
		{"test.one+two@example.com", true},
		{"test@subdomain.example.com", true},
		{"test@example.co.uk", true},
		{"test@example.travel", true},
	}

	invalidCases := []struct {
		input    string
		expected bool
	}{
		{"not an email address", false},
		{"test@example.", false},
		{"@example.com", false},
	}

	// Iterate over the valid test cases
	for _, tc := range validCases {
		result := ContainsEmailAddress(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsEmailAddress(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}

	// Iterate over the invalid test cases
	for _, tc := range invalidCases {
		result := ContainsEmailAddress(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsEmailAddress(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}
}

func TestContainsIP(t *testing.T) {
	validCases := []struct {
		input    string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"255.255.255.255", true},
	}

	invalidCases := []struct {
		input    string
		expected bool
	}{
		{"not an IP address", false},
		{"192.168.1", false},
		{"192.168.1.1.1", false},
		{"192.168.1.", false},
		{"192.168.1.-1", false},
	}

	// Iterate over the valid test cases
	for _, tc := range validCases {
		result := ContainsIP(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsIP(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}

	// Iterate over the invalid test cases
	for _, tc := range invalidCases {
		result := ContainsIP(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsIP(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}
}

func TestContainsFQDN(t *testing.T) {
	validCases := []struct {
		input    string
		expected bool
	}{
		{"example.com", true},
		{"subdomain.example.com", true},
		{"www.example.com", true},
		{"mail.google.com", true},
		{"my.site123.info", true},
	}

	invalidCases := []struct {
		input    string
		expected bool
	}{
		{"not a valid FQDN", false},
		{"example", false},
		{"example.c", false},
		{"example-.com", false},
		{"example._com", false},
		{"example..com", false},
	}

	// Iterate over the valid test cases
	for _, tc := range validCases {
		result := ContainsFQDN(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsFQDN(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}

	// Iterate over the invalid test cases
	for _, tc := range invalidCases {
		result := ContainsFQDN(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsFQDN(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}
}

func TestIsNamespace(t *testing.T) {
	testCases := []struct {
		name     string
		ns       string
		expected bool
	}{
		{
			name:     "valid namespace with two parts",
			ns:       "mycompany.myservice",
			expected: true,
		},
		{
			name:     "valid namespace with three parts",
			ns:       "mycompany.myservice.myenv",
			expected: true,
		},
		{
			name:     "invalid namespace with numeric characters",
			ns:       "mycompany.1234",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNamespace(tc.ns)
			if result != tc.expected {
				t.Errorf("Expected IsNamespace(%q) to be %v, but got %v", tc.ns, tc.expected, result)
			}
		})
	}
}

func TestIsSSN(t *testing.T) {
	validCases := []struct {
		input    string
		expected bool
	}{
		{"123-45-6789", true},
		{"111-22-3333", true},
	}

	invalidCases := []struct {
		input    string
		expected bool
	}{
		{"not an SSN", false},
		{"123-4-6789", false},
		{"123-45-67890", false},
		{"1234-56-7890", false},
		{"1234567890", false},
		{"ABC-DE-FGHI", false},
		{"123-45-6ABC", false},
	}

	// Iterate over the valid test cases
	for _, tc := range validCases {
		result := IsSSN(tc.input)

		if result != tc.expected {
			t.Errorf("IsSSN(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}

	// Iterate over the invalid test cases
	for _, tc := range invalidCases {
		result := IsSSN(tc.input)

		if result != tc.expected {
			t.Errorf("IsSSN(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}
}

func TestContainsPhoneNo(t *testing.T) {
	validCases := []struct {
		input    string
		expected bool
	}{
		{"1234567890", true},
		{"123-456-7890", true},
		{"(123) 456-7890", true},
		{"+1 123-456-7890", true},
		{"+91 1234567890", true},
		{"+1 (123) 456-7890", true},
		{"+1 1234567890", true},
		{"+86 13912345678", true},
	}

	invalidCases := []struct {
		input    string
		expected bool
	}{
		{"not a phone number", false},
		{"123-4567", false},
		{"123-45-6789", false},
		{"(123)-456-7890", false},
		{"+1 1234567", false},
		{"011-123456", false},
	}

	// Iterate over the valid test cases
	for _, tc := range validCases {
		result := ContainsPhoneNo(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsPhoneNo(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}

	// Iterate over the invalid test cases
	for _, tc := range invalidCases {
		result := ContainsPhoneNo(tc.input)

		if result != tc.expected {
			t.Errorf("ContainsPhoneNo(%v) = %v; want %v", tc.input, result, tc.expected)
		}
	}
}

func TestCheckLuhn(t *testing.T) {
	cases := []struct {
		card     string
		expected bool
	}{
		{"4111111111111111", true},
		{"4111111111111", false},
		{"4012888888881881", true},
		{"378282246310005", true},
		{"6011111111111117", true},
		{"5105105105105100", true},
		{"5105105105105106", false},
		{"1234567812345670", true},
		{"1234567812345678", false},
		{"0000000000000000", true},
		{"0000000000000010", false},
	}

	for _, c := range cases {
		got := CheckLuhn(c.card)
		if got != c.expected {
			t.Errorf("CheckLuhn(%q) == %v, expected %v", c.card, got, c.expected)
		}
	}
}
