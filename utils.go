/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * utils.go
 */

package hatchet

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	MAX_SIZE  = 64
	TAIL_SIZE = 7
)

// ToFloat64 converts to float64
func ToFloat64(num interface{}) float64 {
	f := fmt.Sprintf("%v", num)
	x, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return 0
	}
	return float64(x)
}

// ToInt converts to int
func ToInt(num interface{}) int {
	f := fmt.Sprintf("%v", num)
	x, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return 0
	}
	return int(x)
}

func replaceSpecialChars(name string) string {
	for _, sep := range []string{"-", ".", " ", ":", ","} {
		name = strings.ReplaceAll(name, sep, "_")
	}
	return name
}

func getHatchetName(logname string) string {
	temp := filepath.Base(logname)
	hatchetName := replaceSpecialChars(temp)
	i := strings.LastIndex(hatchetName, "_log")
	if i >= 0 && i >= len(temp)-len(".log.gz") {
		hatchetName = replaceSpecialChars(hatchetName[0:i])
	}
	if len(hatchetName) > MAX_SIZE {
		hatchetName = hatchetName[:MAX_SIZE-TAIL_SIZE]
	}
	if i = strings.LastIndex(hatchetName, "_gz"); i > 0 {
		hatchetName = hatchetName[:i]
	}
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, TAIL_SIZE)
	rand.Read(b)
	tail := fmt.Sprintf("%x", b)[:TAIL_SIZE-1]

	r := []rune(hatchetName) // convert string to runes
	if unicode.IsDigit(r[0]) {
		hatchetName = "_" + hatchetName
	}
	return fmt.Sprintf("%v_%v", hatchetName, tail)
}

func EscapeString(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z"}
	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}
	return value
}

func GetSQLDateSubString(start string, end string) string {
	var err error
	substr := "SUBSTR(date, 1, 16)"
	if len(start) < 16 || len(end) < 16 {
		return substr
	}
	var stime, etime time.Time
	layout := "2006-01-02T15:04"
	if stime, err = time.Parse(layout, start[:16]); err != nil {
		return substr
	}
	if etime, err = time.Parse(layout, end[:16]); err != nil {
		return substr
	}
	minutes := etime.Sub(stime).Minutes()
	if minutes < 1 {
		return "SUBSTR(date, 1, 19)"
	} else if minutes < 10 {
		return "SUBSTR(date, 1, 18)||'9'"
	} else if minutes < 60 {
		return "SUBSTR(date, 1, 16)||':59'"
	} else {
		return "SUBSTR(date, 1, 15)||'9:59'"
	}
}

func GetMongoDateSubString(start string, end string) bson.M {
	var err error
	substr := bson.M{"$substr": bson.A{"$date", 0, 10}}
	if len(start) < 16 || len(end) < 16 {
		return substr
	}
	var stime, etime time.Time
	layout := "2006-01-02T15:04"
	if stime, err = time.Parse(layout, start[:16]); err != nil {
		return substr
	}
	if etime, err = time.Parse(layout, end[:16]); err != nil {
		return substr
	}
	minutes := etime.Sub(stime).Minutes()
	if minutes < 1 {
		return bson.M{"$substr": bson.A{"$date", 0, 19}}
	} else if minutes < 10 {
		return bson.M{"$substr": bson.A{"$date", 0, 18}}
	} else if minutes < 60 {
		return bson.M{"$substr": bson.A{"$date", 0, 16}}
	} else {
		return bson.M{"$substr": bson.A{"$date", 0, 15}}
	}
}

func GetHatchetSummary(info HatchetInfo) string {
	arr := []string{}
	if info.Module == "" {
		info.Module = "community"
	}
	if info.Version != "" {
		arr = append(arr, fmt.Sprintf(": MongoDB v%v (%v)", info.Version, info.Module))
	}
	if info.OS != "" {
		arr = append(arr, "os: "+info.OS)
	}
	if info.Arch != "" {
		arr = append(arr, "arch: "+info.Arch)
	}
	return info.Name + strings.Join(arr, ", ")
}

// GetOffsetLimit returns offset, limit
func GetOffsetLimit(str string) (int, int) {
	toks := strings.Split(str, ",")
	if len(toks) >= 2 {
		return ToInt(toks[0]), ToInt(toks[1])
	} else if len(toks) == 1 {
		return 0, ToInt(toks[0])
	}
	return 0, 0
}

func getDateTimeStr(tm time.Time) string {
	dt := tm.Format("2006-01-02T15:04:05.000-0000")
	return dt
}

func GetBufioReader(data []byte) (*bufio.Reader, error) {
	isGzipped := false
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		isGzipped = true
	}

	if isGzipped {
		gzipReader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()

		var buf bytes.Buffer
		if _, err = buf.ReadFrom(gzipReader); err != nil {
			return nil, err
		}

		return bufio.NewReader(&buf), nil
	}

	return bufio.NewReader(bytes.NewReader(data)), nil
}

func ContainsCreditCardNo(card string) bool {
	cardNo := []byte{}
	for i := range card {
		if card[i] >= '0' && card[i] <= '9' {
			cardNo = append(cardNo, card[i])
		}
	}
	matched, _ := regexp.MatchString(`(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9]{2})[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11})`, string(cardNo))
	return matched && CheckLuhn(string(cardNo))
}

func ContainsEmailAddress(email string) bool {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	return emailRegex.MatchString(email)
}

func ContainsIP(ip string) bool {
	ipRegex := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	octets := strings.Split(ip, ".")
	return len(octets) == 4 && ipRegex.MatchString(ip)
}

func ContainsFQDN(fqdn string) bool {
	if i := strings.Index(fqdn, " "); i >= 0 {
		return false
	}
	parts := strings.Split(fqdn, ".")
	if len(parts) < 2 {
		return false
	}
	fqdnRegex := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]\.)+[a-zA-Z]{2,63}`)
	return fqdnRegex.MatchString(fqdn)
}

func IsNamespace(ns string) bool {
	parts := strings.Split(ns, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, part := range parts {
		re := regexp.MustCompile("[^0-9]")
		if !re.MatchString(part) {
			return false
		}
	}
	nsRegex := regexp.MustCompile(`^[^\d][^$.\n\s@]*\.[^.\n\s@]*([.][^.\n\s@]*)?$`)
	return nsRegex.MatchString(ns)
}

func IsSSN(s string) bool {
	ssnRegex := regexp.MustCompile(`\d{3}-\d{2}-\d{4}`)
	digits := strings.ReplaceAll(s, "-", "")
	return len(digits) == 9 && ssnRegex.MatchString(s)
}

func ContainsPhoneNo(phoneNo string) bool {
	re := regexp.MustCompile("[a-zA-Z]")
	if re.MatchString(phoneNo) {
		return false
	}
	re = regexp.MustCompile("[^0-9+]+")
	digits := re.ReplaceAllString(phoneNo, "")
	if (strings.HasPrefix(digits, "+") && len(digits) > 14) || (!strings.HasPrefix(digits, "+") && len(digits) > 11) {
		return false
	}
	phoneRegex := regexp.MustCompile(`(?:\+?\d{1,3}[- ]?)?\d{10,14}|(\+\d{1,3}\s?)?\(\d{3}\)\s?\d{3}[- ]?\d{4}|\d{3}[- ]?\d{3}[- ]?\d{4}`)
	return phoneRegex.MatchString(phoneNo)
}

func CheckLuhn(card string) bool {
	var sum int
	var digit int
	var even bool
	for i := len(card) - 1; i >= 0; i-- {
		digit, _ = strconv.Atoi(string(card[i]))
		if even {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		even = !even
	}
	return sum%10 == 0
}

func ObfuscateWord(word string) string {
	length := len(word)
	lowers := []rune("abcdefghijklmnopqrstuvwxyz")
	uppers := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	digits := []rune("0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range word {
		if unicode.IsLower(rune(word[i])) {
			b[i] = lowers[rand.Intn(len(lowers))]
		} else if unicode.IsUpper(rune(word[i])) {
			b[i] = uppers[rand.Intn(len(uppers))]
		} else if unicode.IsDigit(rune(word[i])) {
			b[i] = digits[rand.Intn(len(digits))]
		} else {
			b[i] = rune(word[i])
		}
	}
	return string(b)
}
