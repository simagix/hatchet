/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * obfuscation.go
 */

package hatchet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

// Pre-compiled regex patterns for obfuscation
var (
	rePort  = regexp.MustCompile(`:\d{2,}`)
	reEmail = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	reIP    = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	reFQDN  = regexp.MustCompile(`([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63}`)
	reNS    = regexp.MustCompile(`[^@$.\n]*\.[^^@.\n]*([.][^^@.\n]*)?`)
	reDigit = regexp.MustCompile("[0-9.]")
	reSSN   = regexp.MustCompile(`\d{3}-\d{2}-\d{4}`)
	reMAC   = regexp.MustCompile(`([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}`)
	reDate  = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)              // ISO date YYYY-MM-DD
	reMRN   = regexp.MustCompile(`(?i)(mrn|acct|id)[:\s#]*\d{6,}`) // Medical Record Number, Account, ID
)

var (
	cities = []string{
		"Atlanta", "Berlin", "Chicago", "Dublin", "ElPaso",
		"Foshan", "Giza", "Hongkong", "Istanbul", "Jakarta",
		"London", "Miami", "NewYork", "Orlando", "Paris",
		"Queens", "Rome", "Sydney", "Taipei", "Utica",
		"Vancouver", "Warsaw", "Xiamen", "Yonkers", "Zurich",
	}
	flowers = []string{
		"Aster", "Begonia", "Carnation", "Daisy", "Erica",
		"Freesia", "Gardenia", "Hyacinth", "Iris", "Jasmine",
		"Kalmia", "Lavender", "Marigold", "Narcissus", "Orchid",
		"Peony", "Rose", "Sunflower", "Tulip", "Ursinia",
		"Violet", "Wisteria", "Xylobium", "Yarrow", "Zinnia",
	}
)

type Obfuscation struct {
	Coefficient float64           `json:"coefficient"`
	DateOffset  int               `json:"date_offset"` // Days to shift dates
	CardMap     map[string]string `json:"card_map"`
	intMap      map[int]int
	IPMap       map[string]string `json:"ip_map"`
	MACMap      map[string]string `json:"mac_map"`
	NameMap     map[string]string `json:"name_map"`
	numberMap   map[string]float64
	PhoneMap    map[string]string `json:"phone_map"`
	SSNMap      map[string]string `json:"ssn_map"`
	IDMap       map[string]string `json:"id_map"` // MRN, Account numbers, etc.
}

func NewObfuscation() *Obfuscation {
	obs := Obfuscation{}
	obs.Coefficient = 0.917 // Fixed coefficient for deterministic results
	obs.DateOffset = -42    // Shift dates back 42 days (deterministic)
	obs.CardMap = make(map[string]string)
	obs.intMap = make(map[int]int)
	obs.IPMap = make(map[string]string)
	obs.MACMap = make(map[string]string)
	obs.NameMap = make(map[string]string)
	obs.numberMap = make(map[string]float64)
	obs.PhoneMap = make(map[string]string)
	obs.SSNMap = make(map[string]string)
	obs.IDMap = make(map[string]string)
	return &obs
}

// hashIndex returns a deterministic index based on the input string
func hashIndex(s string, max int) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32()) % max
}

// hashOctet returns a deterministic octet (0-255) based on input and position
func hashOctet(s string, pos int) int {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%s:%d", s, pos)))
	return int(h.Sum32()) % 256
}

// PrintMappings outputs all obfuscation mappings to stderr in JSON format
func (ptr *Obfuscation) PrintMappings() {
	// Filter out self-mappings from NameMap
	filteredNameMap := make(map[string]string)
	for k, v := range ptr.NameMap {
		if k != v {
			filteredNameMap[k] = v
		}
	}

	mappings := map[string]interface{}{
		"coefficient": ptr.Coefficient,
		"date_offset": ptr.DateOffset,
		"ip_map":      ptr.IPMap,
		"mac_map":     ptr.MACMap,
		"name_map":    filteredNameMap,
		"ssn_map":     ptr.SSNMap,
		"phone_map":   ptr.PhoneMap,
		"card_map":    ptr.CardMap,
		"id_map":      ptr.IDMap,
	}
	data, _ := json.MarshalIndent(mappings, "", "  ")
	fmt.Fprintln(os.Stderr, string(data))
}

func (ptr *Obfuscation) ObfuscateFile(filename string) error {
	return ptr.ObfuscateFileToWriter(filename, os.Stdout)
}

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
			obfuscatedValue = ptr.ObfuscateString(value)
		case int, int32, int64:
			obfuscatedValue = ptr.ObfuscateInt(value)
		case float32, float64:
			obfuscatedValue = ptr.ObfuscateNumber(value)
		default:
			obfuscatedValue = value
		}
		obfuscated = append(obfuscated, bson.E{Key: elem.Key, Value: obfuscatedValue})
	}
	return obfuscated
}

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
			obfuscatedValue = ptr.ObfuscateString(value)
		case int, int32, int64:
			obfuscatedValue = ptr.ObfuscateInt(value)
		case float32, float64:
			obfuscatedValue = ptr.ObfuscateNumber(value)
		default:
			obfuscatedValue = value
		}
		obfuscated = append(obfuscated, obfuscatedValue)
	}
	return obfuscated
}

// ObfuscateInt uses the original value times the coefficient
func (ptr *Obfuscation) ObfuscateInt(data interface{}) int {
	value := ToInt(data)
	if value <= 1 { // this can be true/false
		return value
	} else if ptr.intMap[value] > 0 {
		return ptr.intMap[value]
	}
	newValue := int(float64(value) * ptr.Coefficient)
	ptr.intMap[value] = newValue
	return newValue
}

// ObfuscateNumber uses the original value times the coefficient
func (ptr *Obfuscation) ObfuscateNumber(data interface{}) float64 {
	value := ToFloat64(data)
	key := fmt.Sprintf("%f", value)
	if ptr.numberMap[key] > 0 {
		return ptr.numberMap[key]
	}
	newValue := float64(value) * ptr.Coefficient
	ptr.numberMap[key] = newValue
	return newValue
}

func (ptr *Obfuscation) ObfuscateString(value string) string {
	matches := rePort.FindStringSubmatch(value)
	if len(matches) > 0 {
		matched := matches[0]
		newValue := fmt.Sprintf(":%v", int(float64(ToInt(matched[1:]))*ptr.Coefficient))
		value = strings.Replace(value, matched, newValue, -1)
	}
	if ContainsCreditCardNo(value) {
		value = ptr.ObfuscateCreditCardNo(value)
	}
	// the following 3, don't change the order
	value = ptr.ObfuscateEmail(value)
	value = ptr.ObfuscateNS(value)
	value = ptr.ObfuscateFQDN(value)
	value = ptr.ObfuscateIP(value)
	value = ptr.ObfuscateMAC(value)
	value = ptr.ObfuscateSSN(value)
	value = ptr.ObfuscatePhoneNo(value)
	value = ptr.ObfuscateDate(value)
	value = ptr.ObfuscateID(value)
	return value
}

// ObfuscateCreditCardNo obfuscate digits with '*' except for last 4 digits
func (ptr *Obfuscation) ObfuscateCreditCardNo(cardNo string) string {
	if !ContainsCreditCardNo(cardNo) {
		return cardNo
	}
	lastFourDigits := cardNo[len(cardNo)-4:]
	obfuscated := make([]rune, len(cardNo)-4)
	for i, c := range cardNo[:len(cardNo)-4] {
		if c >= '0' && c <= '9' {
			if c == ' ' || c == '-' {
				obfuscated[i] = c
			} else {
				obfuscated[i] = '*'
			}
		} else {
			obfuscated[i] = c
		}
	}
	return string(obfuscated) + lastFourDigits
}

// ObfuscateEmail replace domain name with a city name and user name with a flower name
func (ptr *Obfuscation) ObfuscateEmail(email string) string {
	if !ContainsEmailAddress(email) {
		return email
	}
	matches := reEmail.FindStringSubmatch(email)
	if len(matches) > 0 {
		matched := matches[0]
		if ptr.NameMap[matched] != "" {
			return strings.Replace(email, matched, ptr.NameMap[matched], -1)
		}
		// Use deterministic selection based on the matched value
		city := cities[hashIndex(matched, len(cities))]
		flower := flowers[hashIndex(matched+"flower", len(flowers))]
		newValue := strings.ToLower(flower + "@" + city + ".com")
		ptr.NameMap[matched] = newValue
		ptr.NameMap[newValue] = newValue
		return strings.Replace(email, matched, newValue, -1)
	}
	return email
}

// ObfuscateIP returns a new IP but keep the first and last octets the same
func (ptr *Obfuscation) ObfuscateIP(ip string) string {
	if !ContainsIP(ip) {
		return ip
	}
	matches := reIP.FindStringSubmatch(ip)
	if len(matches) > 0 {
		matched := matches[0]
		if matched == "0.0.0.0" || matched == "127.0.0.1" {
			return ip
		}
		newValue := ""
		if ptr.IPMap[matched] != "" {
			newValue = ptr.IPMap[matched]
		} else {
			octets := strings.Split(matched, ".")
			if len(octets) != 4 {
				return ip
			}
			// Use deterministic octet generation based on the IP
			newValue = octets[0] + "." + strconv.Itoa(hashOctet(matched, 1)) + "." + strconv.Itoa(hashOctet(matched, 2)) + "." + octets[3]
			ptr.IPMap[matched] = newValue
		}
		return strings.Replace(ip, matched, newValue, -1)
	}
	return ip
}

// ObfuscateFQDN returns a obfuscated FQDN
func (ptr *Obfuscation) ObfuscateFQDN(fqdn string) string {
	// Skip file paths - they're not FQDNs
	if strings.Contains(fqdn, "/") || strings.Contains(fqdn, "\\") {
		return fqdn
	}
	if !ContainsFQDN(fqdn) {
		return fqdn
	}
	matches := reFQDN.FindStringSubmatch(fqdn)
	if len(matches) > 0 {
		matched := matches[0]
		if ptr.NameMap[matched] != "" {
			return strings.Replace(fqdn, matched, ptr.NameMap[matched], -1)
		}
		newValue := ptr.generateObfuscatedName(matched)
		ptr.NameMap[matched] = newValue
		ptr.NameMap[newValue] = newValue // so, it won't be replaced again
		return strings.Replace(fqdn, matched, newValue, -1)
	}
	return fqdn
}

// ObfuscateNS returns a obfuscated namespace
func (ptr *Obfuscation) ObfuscateNS(ns string) string {
	// Skip file paths - they're not namespaces
	if strings.Contains(ns, "/") || strings.Contains(ns, "\\") {
		return ns
	}
	if !IsNamespace(ns) {
		return ns
	}
	charts := reDigit.ReplaceAllString(ns, "")
	if len(charts) == 0 {
		return ns
	}
	matches := reNS.FindStringSubmatch(ns)
	if len(matches) > 0 {
		matched := matches[0]
		if ptr.NameMap[matched] != "" {
			return strings.Replace(ns, matched, ptr.NameMap[matched], -1)
		}
		newValue := ptr.generateObfuscatedName(matched)
		ptr.NameMap[matched] = newValue
		ptr.NameMap[newValue] = newValue
		return strings.Replace(ns, matched, newValue, -1)
	}
	return ns
}

// generateObfuscatedName generates an obfuscated name from city and flower names
func (ptr *Obfuscation) generateObfuscatedName(matched string) string {
	// Use deterministic selection based on the matched value
	city := cities[hashIndex(matched, len(cities))]
	flower := flowers[hashIndex(matched+"flower", len(flowers))]
	parts := strings.Split(matched, ".")
	if len(parts) > 2 {
		tail := parts[len(parts)-1]
		return strings.ToLower(flower + "." + city + "." + tail)
	}
	return strings.ToLower(city + "." + flower)
}

// ObfuscateSSN returns a obfuscated SSN
func (ptr *Obfuscation) ObfuscateSSN(ssn string) string {
	if !IsSSN(ssn) {
		return ssn
	}
	matches := reSSN.FindStringSubmatch(ssn)
	if len(matches) > 0 {
		matched := matches[0]
		newValue := ""
		if ptr.SSNMap[matched] != "" {
			newValue = ptr.SSNMap[matched]
		} else {
			digits := []byte{}
			for _, c := range matched {
				if c >= '0' && c <= '9' {
					digits = append(digits, byte(c))
				}
			}
			// Deterministic shuffle using hash
			for i := len(digits) - 1; i > 0; i-- {
				j := hashIndex(matched+strconv.Itoa(i), i+1)
				digits[i], digits[j] = digits[j], digits[i]
			}
			newValue = string(digits[:3]) + "-" + string(digits[3:5]) + "-" + string(digits[5:])
			ptr.SSNMap[matched] = newValue
		}
		return strings.Replace(ssn, matched, newValue, -1)
	}
	return ssn
}

// ObfuscatePhoneNo returns a deterministic obfuscated phone number with the same format
func (ptr *Obfuscation) ObfuscatePhoneNo(phoneNo string) string {
	if !ContainsPhoneNo(phoneNo) {
		return phoneNo
	}
	if ptr.PhoneMap[phoneNo] != "" {
		return ptr.PhoneMap[phoneNo]
	}
	obfuscated := make([]byte, len(phoneNo))
	n := 0
	for i := range obfuscated {
		if phoneNo[i] >= '0' && phoneNo[i] <= '9' {
			n++
			if n > 5 {
				// Use deterministic digit based on phone number and position
				obfuscated[i] = byte(hashIndex(phoneNo+strconv.Itoa(i), 10) + '0')
			} else {
				obfuscated[i] = phoneNo[i]
			}
		} else {
			obfuscated[i] = phoneNo[i]
		}
	}
	ptr.PhoneMap[phoneNo] = string(obfuscated)
	return string(obfuscated)
}

// ObfuscateMAC returns a deterministic obfuscated MAC address
func (ptr *Obfuscation) ObfuscateMAC(value string) string {
	matches := reMAC.FindStringSubmatch(value)
	if len(matches) == 0 {
		return value
	}
	matched := matches[0]
	if ptr.MACMap[matched] != "" {
		return strings.Replace(value, matched, ptr.MACMap[matched], -1)
	}
	// Determine separator used (: or -)
	sep := ":"
	if strings.Contains(matched, "-") {
		sep = "-"
	}
	// Parse original octets
	parts := strings.FieldsFunc(matched, func(r rune) bool { return r == ':' || r == '-' })
	if len(parts) != 6 {
		return value
	}
	// Keep vendor prefix (first 3 octets), obfuscate device ID (last 3 octets)
	newParts := make([]string, 6)
	copy(newParts[:3], parts[:3])
	for i := 3; i < 6; i++ {
		newParts[i] = fmt.Sprintf("%02X", hashOctet(matched, i))
	}
	newValue := strings.Join(newParts, sep)
	ptr.MACMap[matched] = newValue
	return strings.Replace(value, matched, newValue, -1)
}

// ObfuscateDate shifts dates by a fixed offset while preserving format
func (ptr *Obfuscation) ObfuscateDate(value string) string {
	matches := reDate.FindAllString(value, -1)
	if len(matches) == 0 {
		return value
	}
	for _, matched := range matches {
		// Parse the date
		year, _ := strconv.Atoi(matched[0:4])
		month, _ := strconv.Atoi(matched[5:7])
		day, _ := strconv.Atoi(matched[8:10])

		// Simple date shift - add offset to day, handle overflow simply
		day += ptr.DateOffset
		for day < 1 {
			month--
			if month < 1 {
				month = 12
				year--
			}
			day += 30 // Approximate
		}
		for day > 28 {
			day -= 28
			month++
			if month > 12 {
				month = 1
				year++
			}
		}

		newValue := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		value = strings.Replace(value, matched, newValue, 1)
	}
	return value
}

// ObfuscateID obfuscates MRN, account numbers, and other IDs
func (ptr *Obfuscation) ObfuscateID(value string) string {
	matches := reMRN.FindStringSubmatch(value)
	if len(matches) == 0 {
		return value
	}
	matched := matches[0]
	if ptr.IDMap[matched] != "" {
		return strings.Replace(value, matched, ptr.IDMap[matched], -1)
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
		newDigits[i] = byte(hashIndex(matched+strconv.Itoa(i), 10) + '0')
	}

	newValue := prefix + string(newDigits)
	ptr.IDMap[matched] = newValue
	return strings.Replace(value, matched, newValue, -1)
}
