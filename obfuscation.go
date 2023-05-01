/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * obfuscation.go
 */

package hatchet

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	cities = []string{
		"Atlanta", "Berlin", "Chicago", "Dublin", "ElPaso",
		"FortWorth", "Greenville", "Hongkong", "Indianapolis", "Jacksonville",
		"London", "Miami", "NewYork", "Orlando", "Paris",
		"Queens", "Rome", "Sydney", "Taipei", "Utica",
		"Vancouver", "Warsaw", "Xiamen", "Yonkers", "Zurich",
	}
	flowers = []string{
		"Aster", "Begonia", "Carnation", "Daisy", "Echinacea",
		"Freesia", "Gardenia", "Hyacinth", "Iris", "Jasmine",
		"Kalmia", "Lavender", "Marigold", "Narcissus", "Orchid",
		"Peony", "Rose", "Sunflower", "Tulip", "Ursinia",
		"Violet", "Wisteria", "Xeranthemum", "Yarrow", "Zinnia",
	}
)

type Obfuscation struct {
	Coefficient float64           `json:"coefficient"`
	CardMap     map[string]string `json:"card_map"`
	EmailMap    map[string]string `json:"email_map"`
	HostMap     map[string]string `json:"host_map"`
	intMap      map[int]int
	IPMap       map[string]string `json:"ip_map"`
	numberMap   map[string]float64
	PhoneMap    map[string]string `json:"phone_map"`
	SSNMap      map[string]string `json:"ssn_map"`
}

func NewObfuscation() *Obfuscation {
	obs := Obfuscation{}
	rand.Seed(time.Now().UnixNano())
	obs.Coefficient = 0.923 - rand.Float64()*0.05
	obs.Coefficient = math.Round(obs.Coefficient*1000) / 1000
	obs.CardMap = make(map[string]string)
	obs.EmailMap = make(map[string]string)
	obs.HostMap = make(map[string]string)
	obs.intMap = make(map[int]int)
	obs.IPMap = make(map[string]string)
	obs.numberMap = make(map[string]float64)
	obs.PhoneMap = make(map[string]string)
	obs.SSNMap = make(map[string]string)
	return &obs
}

func (ptr *Obfuscation) ObfuscateFile(filename string) error {
	var buf []byte
	var isPrefix bool
	var reader *bufio.Reader
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if reader, err = gox.NewReader(file); err != nil {
		return err
	}

	for {
		if buf, isPrefix, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
			break
		}
		if len(buf) == 0 {
			continue
		}
		str := string(buf)
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
		attr, ok := doc.Map()["attr"].(bson.D)
		if !ok {
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
			fmt.Println(string(buf))
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
	if ptr.intMap[value] > 0 {
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
	portRegex := regexp.MustCompile(`:\d{2,}`)
	matches := portRegex.FindStringSubmatch(value)
	if len(matches) > 0 {
		matched := matches[0]
		newValue := fmt.Sprintf(":%v", int(float64(ToInt(matched[1:]))*ptr.Coefficient))
		value = strings.Replace(value, matched, newValue, -1)
	}
	if IsCreditCardNo(value) {
		value = ptr.ObfuscateCreditCardNo(value)
	}
	if IsFQDN(value) {
		value = ptr.ObfuscateFQDN(value)
	}
	if IsEmail(value) {
		value = ptr.ObfuscateEmail(value)
	}
	if IsIP(value) {
		value = ptr.ObfuscateIP(value)
	}
	if IsSSN(value) {
		value = ptr.ObfuscateSSN(value)
	}
	if IsPhoneNo(value) {
		value = ptr.ObfuscatePhoneNo(value)
	}
	return value
}

// ObfuscateCreditCardNo obfuscate digits with '*' except for last 4 digits
func (ptr *Obfuscation) ObfuscateCreditCardNo(cardNo string) string {
	if len(cardNo) < 4 {
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
	rand.Seed(time.Now().UnixNano())
	city := cities[rand.Intn(len(cities))]
	flower := flowers[rand.Intn(len(flowers))]
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	matches := emailRegex.FindStringSubmatch(email)
	if len(matches) > 0 {
		matched := matches[0]
		newValue := ""
		if ptr.EmailMap[matched] != "" {
			newValue = ptr.EmailMap[matched]
		} else {
			newValue = strings.ToLower(flower + "@" + city + ".com")
			ptr.EmailMap[matched] = newValue
		}
		return strings.Replace(email, matched, newValue, -1)
	}
	return email
}

// ObfuscateIP returns a new IP but keep the first and last octets the same
func (ptr *Obfuscation) ObfuscateIP(ip string) string {
	ipRegex := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	matches := ipRegex.FindStringSubmatch(ip)
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
			newValue = octets[0] + "." + strconv.Itoa(rand.Intn(256)) + "." + strconv.Itoa(rand.Intn(256)) + "." + octets[3]
			ptr.IPMap[matched] = newValue
		}
		return strings.Replace(ip, matched, newValue, -1)
	}
	return ip
}

func (ptr *Obfuscation) ObfuscateFQDN(fqdn string) string {
	rand.Seed(time.Now().UnixNano())
	fqdnRegex := regexp.MustCompile(`([_a-zA-Z0-9]+(-[_a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}`)
	matches := fqdnRegex.FindStringSubmatch(fqdn)
	if len(matches) > 0 {
		matched := matches[0]
		newValue := ""
		if ptr.HostMap[matched] != "" {
			newValue = ptr.HostMap[matched]
		} else {
			city := cities[rand.Intn(len(cities))]
			flower := flowers[rand.Intn(len(flowers))]
			parts := strings.Split(matches[0], ".")
			if len(parts) > 2 {
				tail := parts[len(parts)-1]
				newValue = strings.ToLower(flower + "." + city + "." + tail)
			} else {
				newValue = strings.ToLower(city + "." + flower)
			}
			ptr.HostMap[matched] = newValue
		}
		return strings.Replace(fqdn, matched, newValue, -1)
	}
	return fqdn
}

// ObfuscateSSN returns a random SSN
func (ptr *Obfuscation) ObfuscateSSN(ssn string) string {
	ssnRegex := regexp.MustCompile(`\d{3}-?\d{2}-?\d{4}`)
	matches := ssnRegex.FindStringSubmatch(ssn)
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
			for i := range digits {
				j := rand.Intn(i + 1)
				digits[i], digits[j] = digits[j], digits[i]
			}
			newValue = string(digits[:3]) + "-" + string(digits[3:5]) + "-" + string(digits[5:])
			ptr.SSNMap[matched] = newValue
		}
		return strings.Replace(ssn, matched, newValue, -1)
	}
	return ssn
}

// ObfuscatePhoneNo returns a randome phone number with the same format
func (ptr *Obfuscation) ObfuscatePhoneNo(phone string) string {
	if ptr.PhoneMap[phone] != "" {
		return ptr.PhoneMap[phone]
	}
	rand.Seed(time.Now().UnixNano())
	obfuscated := make([]byte, len(phone))
	n := 0
	for i := range obfuscated {
		if phone[i] >= '0' && phone[i] <= '9' {
			n++
			if n > 5 {
				obfuscated[i] = byte(rand.Intn(10) + '0')
			} else {
				obfuscated[i] = phone[i]
			}
		} else {
			obfuscated[i] = phone[i]
		}
	}
	ptr.PhoneMap[phone] = string(obfuscated)
	return string(obfuscated)
}
