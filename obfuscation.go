/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * obfuscation.go
 */

package hatchet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

// Obfuscation wraps gox.Obfuscator with hatchet-specific functionality
type Obfuscation struct {
	*gox.Obfuscator
}

// NewObfuscation creates a new Obfuscation instance
func NewObfuscation() *Obfuscation {
	return &Obfuscation{
		Obfuscator: gox.NewObfuscator(),
	}
}

// PrintMappings outputs all obfuscation mappings to stderr in JSON format
func (ptr *Obfuscation) PrintMappings() {
	mappings := ptr.Obfuscator.GetMappings()
	data, _ := json.MarshalIndent(mappings, "", "  ")
	fmt.Fprintln(os.Stderr, string(data))
}

// ObfuscateFile reads a log file and writes obfuscated output to stdout
func (ptr *Obfuscation) ObfuscateFile(filename string) error {
	return ptr.ObfuscateFileToWriter(filename, os.Stdout)
}

// ObfuscateFileToWriter reads a log file and writes obfuscated output to the provided writer
func (ptr *Obfuscation) ObfuscateFileToWriter(filename string, w io.Writer) error {
	var err error
	var buf []byte
	var isPrefix bool
	var reader *bufio.Reader
	var scanner *bufio.Scanner

	if filename == "-" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		if reader, err = gox.NewReader(file); err != nil {
			return err
		}
		scanner = bufio.NewScanner(reader)
	}

	for scanner.Scan() {
		str := scanner.Text()
		if str == "" {
			continue
		}
		for isPrefix {
			var bbuf []byte
			if bbuf, isPrefix, err = reader.ReadLine(); err != nil {
				break
			}
			str += string(bbuf)
		}
		var doc bson.D
		if err = bson.UnmarshalExtJSON([]byte(str), false, &doc); err != nil {
			continue
		}
		// Find the attr element directly from bson.D to preserve type
		var attr bson.D
		for _, elem := range doc {
			if elem.Key == "attr" {
				if a, ok := elem.Value.(bson.D); ok {
					attr = a
				}
				break
			}
		}
		if attr == nil {
			continue
		}
		obfuscated := ptr.ObfuscateBsonD(attr)
		document := bson.D{}
		for _, elem := range doc {
			if elem.Key != "attr" {
				document = append(document, elem)
			} else if len(obfuscated) > 0 {
				document = append(document, bson.E{Key: "attr", Value: obfuscated})
			}
		}
		buf, err = bson.MarshalExtJSON(document, false, false)
		if err == nil {
			fmt.Fprintln(w, string(buf))
		}
	}
	return nil
}

// ObfuscateBsonD recursively obfuscates a bson.D document while preserving key order
func (ptr *Obfuscation) ObfuscateBsonD(d bson.D) bson.D {
	var obfuscated bson.D
	for _, elem := range d {
		var obfuscatedValue interface{}
		switch value := elem.Value.(type) {
		case bson.D:
			obfuscatedValue = ptr.ObfuscateBsonD(value)
		case bson.A:
			obfuscatedValue = ptr.ObfuscateBsonA(value)
		case string:
			obfuscatedValue = ptr.Obfuscator.ObfuscateString(value)
		case int:
			obfuscatedValue = ptr.Obfuscator.ObfuscateInt(value)
		case int32:
			obfuscatedValue = int32(ptr.Obfuscator.ObfuscateInt(int(value)))
		case int64:
			obfuscatedValue = int64(ptr.Obfuscator.ObfuscateInt(int(value)))
		case float32:
			obfuscatedValue = float32(ptr.Obfuscator.ObfuscateNumber(float64(value)))
		case float64:
			obfuscatedValue = ptr.Obfuscator.ObfuscateNumber(value)
		default:
			obfuscatedValue = value
		}
		obfuscated = append(obfuscated, bson.E{Key: elem.Key, Value: obfuscatedValue})
	}
	return obfuscated
}

// ObfuscateBsonA recursively obfuscates a bson.A array
func (ptr *Obfuscation) ObfuscateBsonA(a bson.A) bson.A {
	var obfuscated bson.A
	for _, elem := range a {
		var obfuscatedValue interface{}
		switch value := elem.(type) {
		case bson.D:
			obfuscatedValue = ptr.ObfuscateBsonD(value)
		case bson.A:
			obfuscatedValue = ptr.ObfuscateBsonA(value)
		case string:
			obfuscatedValue = ptr.Obfuscator.ObfuscateString(value)
		case int:
			obfuscatedValue = ptr.Obfuscator.ObfuscateInt(value)
		case int32:
			obfuscatedValue = int32(ptr.Obfuscator.ObfuscateInt(int(value)))
		case int64:
			obfuscatedValue = int64(ptr.Obfuscator.ObfuscateInt(int(value)))
		case float32:
			obfuscatedValue = float32(ptr.Obfuscator.ObfuscateNumber(float64(value)))
		case float64:
			obfuscatedValue = ptr.Obfuscator.ObfuscateNumber(value)
		default:
			obfuscatedValue = value
		}
		obfuscated = append(obfuscated, obfuscatedValue)
	}
	return obfuscated
}

// --- Wrapper methods for backward compatibility ---

// ObfuscateInt obfuscates an integer value
func (ptr *Obfuscation) ObfuscateInt(data interface{}) int {
	return ptr.Obfuscator.ObfuscateInt(gox.ToInt(data))
}

// ObfuscateNumber obfuscates a numeric value
func (ptr *Obfuscation) ObfuscateNumber(data interface{}) float64 {
	f, _ := gox.ToFloat64(data)
	return ptr.Obfuscator.ObfuscateNumber(f)
}

// ObfuscateCreditCardNo obfuscates a credit card number
func (ptr *Obfuscation) ObfuscateCreditCardNo(cardNo string) string {
	if !ContainsCreditCardNo(cardNo) {
		return cardNo
	}
	return ptr.Obfuscator.ObfuscateCreditCardNo(cardNo)
}

// ObfuscateEmail obfuscates an email address
func (ptr *Obfuscation) ObfuscateEmail(email string) string {
	return ptr.Obfuscator.ObfuscateEmail(email)
}

// ObfuscateIP obfuscates an IP address
func (ptr *Obfuscation) ObfuscateIP(ip string) string {
	return ptr.Obfuscator.ObfuscateIP(ip)
}

// ObfuscateFQDN obfuscates a fully qualified domain name
func (ptr *Obfuscation) ObfuscateFQDN(fqdn string) string {
	return ptr.Obfuscator.ObfuscateFQDN(fqdn)
}

// ObfuscateNS obfuscates a MongoDB namespace (db.collection)
// Uses hatchet's stricter IsNamespace check which excludes email addresses
func (ptr *Obfuscation) ObfuscateNS(ns string) string {
	// Use hatchet's IsNamespace which excludes emails (has @ check)
	if !IsNamespace(ns) {
		return ns
	}
	return ptr.Obfuscator.ObfuscateNamespace(ns)
}

// ObfuscateSSN obfuscates a Social Security Number
func (ptr *Obfuscation) ObfuscateSSN(ssn string) string {
	return ptr.Obfuscator.ObfuscateSSN(ssn)
}

// ObfuscatePhoneNo obfuscates a phone number
func (ptr *Obfuscation) ObfuscatePhoneNo(phoneNo string) string {
	return ptr.Obfuscator.ObfuscatePhoneNo(phoneNo)
}

// ObfuscateMAC obfuscates a MAC address
func (ptr *Obfuscation) ObfuscateMAC(mac string) string {
	return ptr.Obfuscator.ObfuscateMAC(mac)
}

// ObfuscateDate obfuscates dates in a string
func (ptr *Obfuscation) ObfuscateDate(value string) string {
	return ptr.Obfuscator.ObfuscateDate(value)
}

// ObfuscateID obfuscates MRN, account numbers, and other IDs
func (ptr *Obfuscation) ObfuscateID(value string) string {
	// Use gox regex and obfuscation logic
	matches := gox.ReMRN.FindStringSubmatch(value)
	if len(matches) == 0 {
		return value
	}
	matched := matches[0]
	if cached, exists := ptr.IDMap[matched]; exists {
		return strings.Replace(value, matched, cached, -1)
	}

	// Find where the number starts
	numStart := 0
	for i, c := range matched {
		if c >= '0' && c <= '9' {
			numStart = i
			break
		}
	}

	prefix := matched[:numStart]
	numPart := matched[numStart:]

	// Generate deterministic replacement digits
	newDigits := make([]byte, len(numPart))
	for i := range numPart {
		newDigits[i] = byte(gox.HashIndex(matched+fmt.Sprintf("%d", i), 10) + '0')
	}

	newValue := prefix + string(newDigits)
	ptr.IDMap[matched] = newValue
	return strings.Replace(value, matched, newValue, -1)
}
