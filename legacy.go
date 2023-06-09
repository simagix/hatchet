/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * legacy.go
 */

package hatchet

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddLegacyString converts log to legacy format
func AddLegacyString(doc *Logv2Info) error {
	var err error
	var arr []string
	attrMap := doc.Attr.Map()

	if doc.Msg != "Slow query" {
		if doc.Msg == "Connection ended" {
			arr = append(arr, "end connection")
		} else if doc.Msg == "Connection accepted" {
			arr = append(arr, "connection accepted")
		} else if doc.Msg == "Authentication succeeded" {
			arr = append(arr, "Successfully authenticated")
		} else {
			arr = append(arr, doc.Msg)
		}
	}

	if doc.Component == "CONTROL" && attrMap["host"] != nil {
		arr = append(arr, fmt.Sprintf("pid=%v port=%v %v host=%v",
			attrMap["pid"], attrMap["port"], attrMap["architecture"], attrMap["host"]))
	} else if doc.Component == "ACCESS" {
		for _, attr := range doc.Attr {
			if attr.Key == "authenticationDatabase" {
				arr = append(arr, fmt.Sprintf("on %v", attr.Value))
			} else if attr.Key == "principalName" {
				arr = append(arr, fmt.Sprintf("as principal %v", attr.Value))
			} else if attr.Key == "remote" {
				arr = append(arr, fmt.Sprintf("from client %v", attr.Value))
			} else if attr.Key == "durationMillis" {
				arr = append(arr, fmt.Sprintf("%vms", attr.Value))
			}
		}
	} else if doc.Component == "NETWORK" {
		remote := RemoteClient{}
		for _, attr := range doc.Attr {
			if attr.Key == "remote" {
				toks := strings.Split(attr.Value.(string), ":")
				remote.IP = toks[0]
				remote.Port = toks[1]
				if doc.Msg == "Connection ended" {
					remote.Ended = 1
					arr = append(arr, fmt.Sprintf("%v", attr.Value))
				} else if doc.Msg == "Connection accepted" {
					remote.Accepted = 1
					arr = append(arr, fmt.Sprintf("from %v", attr.Value))
				} else {
					arr = append(arr, fmt.Sprintf(`"%v":"%v"`, attr.Key, attr.Value))
				}
			} else if attr.Key == "client" {
				arr = append(arr, fmt.Sprintf(`"%v":"%v"`, attr.Key, attr.Value))
			} else if attr.Key == "connectionId" { // && doc.Msg != "Connection ended" {
				arr = append(arr, fmt.Sprintf("#%v", attr.Value))
			} else if attr.Key == "connectionCount" {
				arr = append(arr, fmt.Sprintf("(%v connections now open)", attr.Value))
				remote.Conns = ToInt(attr.Value)
			} else if attr.Key == "doc" {
				b, _ := bson.MarshalExtJSON(attr.Value, false, false)
				arr = append(arr, fmt.Sprintf(`"%v":"%v"`, attr.Key, string(b)))
				if doc.Msg == "client metadata" {
					data, ok := attr.Value.(bson.D)
					if ok {
						driver, ok := data.Map()["driver"].(bson.D)
						if ok {
							remote.Driver, _ = driver.Map()["name"].(string)
							remote.Version, _ = driver.Map()["version"].(string)
						}
					}
				}
			} else {
				b, err := bson.MarshalExtJSON(attr.Value, false, false)
				if err != nil {
					arr = append(arr, fmt.Sprintf(`%v:"%v"`, attr.Key, attr.Value))
				} else {
					arr = append(arr, fmt.Sprintf(`%v:"%v"`, attr.Key, string(b)))
				}
			}
		}
		if remote.IP != "" {
			doc.Client = &remote
		}
	} else {
		for _, attr := range doc.Attr {
			if attr.Key == "type" || attr.Key == "ns" {
				str := attr.Value.(string)
				arr = append(arr, str)
			} else if attr.Key == "durationMillis" {
				arr = append(arr, fmt.Sprintf("%vms", attr.Value))
			} else {
				arr = append(arr, fmt.Sprintf("%v:%v", attr.Key, toLegacyString(attr.Value)))
			}

			// extra effort of retrieving driver info from COMMAND
			if attr.Key == "command" {
				remote := RemoteClient{}
				if _, ok := attr.Value.(bson.D); !ok {
					continue
				}
				_client, ok := attr.Value.(bson.D).Map()["$client"].(bson.D)
				if ok {
					driver, ok := _client.Map()["driver"].(bson.D)
					if ok {
						remote.Driver, _ = driver.Map()["name"].(string)
						remote.Version, _ = driver.Map()["version"].(string)
					}
					mongos, ok := _client.Map()["mongos"].(bson.D)
					if ok {
						remote.IP, ok = mongos.Map()["client"].(string)
						if ok {
							if strings.Contains(remote.IP, ":") {
								toks := strings.Split(remote.IP, ":")
								remote.IP = toks[0]
							}
						}
					}
				}
				if remote.IP != "" {
					doc.Client = &remote
				}
			}
		}
	}

	if len(arr) == 0 {
		return nil
	}
	doc.Message = strings.Join(arr, " ")
	return err
}

func toLegacyString(o interface{}) interface{} {
	switch data := o.(type) {
	case nil:
		return o
	case bool:
		return fmt.Sprintf(" %v", o)
	case bson.A:
		arrays := []string{}
		for _, list := range data {
			arr := []string{}
			if _, ok := list.(bson.D); ok {
				for _, doc := range list.(bson.D) {
					if doc.Key == "" {
						doc.Key = `""`
					}
					arr = append(arr, fmt.Sprintf("{ %v:%v }", doc.Key, toLegacyString(doc.Value)))
				}
			} else {
				arr = append(arr, fmt.Sprintf("%v", toLegacyString(list)))
			}
			arrays = append(arrays, strings.Join(arr, ", "))
		}
		return "[" + strings.Join(arrays, ", ") + "]"
	case bson.D:
		arr := []string{}
		for _, doc := range data {
			if doc.Key == "" {
				doc.Key = `""`
			}
			arr = append(arr, fmt.Sprintf("%v:%v", doc.Key, toLegacyString(doc.Value)))
		}
		return " { " + strings.Join(arr, ", ") + " }"
	case bson.E:
		val := toLegacyString(data.Value)
		if strings.Index(data.Key, ".") > 0 {
			return fmt.Sprintf(` { "%v":%v } `, data.Key, val)
		}
		if data.Key == "" {
			data.Key = `""`
		}
		return fmt.Sprintf(" { %v:%v } ", data.Key, val)
	case int, int32, int64, float32, float64, primitive.Decimal128:
		return o
	case primitive.Binary:
		if data.Subtype == 0 {
			x := base64.StdEncoding.EncodeToString(data.Data)
			return fmt.Sprintf(`{ $binary:{ base64: "%v", subtype:0}}`, x)
		} else if data.Subtype == 4 {
			x := hex.EncodeToString(data.Data)
			return fmt.Sprintf(`{ $uuid: "%s-%s-%s-%s-%s"}`, x[:8], x[8:12], x[12:16], x[16:20], x[20:])
		} else {
			log.Println("unhandled subtype", data.Subtype)
		}
	case primitive.ObjectID:
		return fmt.Sprintf(`{ $oid: "%v"}`, data.Hex())
	case primitive.Timestamp:
		return fmt.Sprintf(`{ t:%v, i:%v}`, data.T, data.I)
	case string, primitive.DateTime:
		return fmt.Sprintf(` "%v"`, o)
	case primitive.Regex:
		return fmt.Sprintf(" /%v/%v", data.Pattern, data.Options)
	case primitive.MinKey, primitive.MaxKey:
		return fmt.Sprintf(` %v`, o)
	default:
		log.Printf("unhandled data type %T, returned original value: %v", o, o)
		return fmt.Sprintf(` %v`, o)
	}
	return o
}
