// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
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
		remote := Remote{}
		for _, attr := range doc.Attr {
			if attr.Key == "remote" {
				toks := strings.Split(attr.Value.(string), ":")
				remote.IP = toks[0]
				remote.Port = toks[1]
				if doc.Msg == "Connection ended" {
					remote.Ended = 1
					arr = append(arr, fmt.Sprintf("%v", attr.Value))
				} else {
					remote.Accepted = 1
					arr = append(arr, fmt.Sprintf("from %v", attr.Value))
				}
			} else if attr.Key == "client" {
				arr = append(arr, fmt.Sprintf("%v:", attr.Value))
			} else if attr.Key == "connectionId" && doc.Msg != "Connection ended" {
				arr = append(arr, fmt.Sprintf("#%v", attr.Value))
			} else if attr.Key == "connectionCount" {
				arr = append(arr, fmt.Sprintf("(%v connections now open)", attr.Value))
				remote.Conns = ToInt(attr.Value)
			} else if attr.Key == "doc" {
				b, _ := bson.MarshalExtJSON(attr.Value, false, false)
				arr = append(arr, string(b))
			}
		}
		if remote.IP != "" {
			doc.Remote = &remote
		}
	} else if doc.Component == "COMMAND" || doc.Component == "WRITE" || doc.Component == "QUERY" || doc.Component == "TXT" {
		for _, attr := range doc.Attr {
			if attr.Key == "type" || attr.Key == "ns" {
				str := attr.Value.(string)
				arr = append(arr, str)
			} else if attr.Key == "durationMillis" {
				arr = append(arr, fmt.Sprintf("%vms", attr.Value))
			} else {
				arr = append(arr, fmt.Sprintf("%v:%v", attr.Key, toLegacyString(attr.Value)))
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
	if list, ok := o.(bson.A); ok {
		arrs := []string{}
		for _, alist := range list {
			arr := []string{}
			if _, ok := alist.(bson.D); ok {
				for _, doc := range alist.(bson.D) {
					arr = append(arr, fmt.Sprintf("{ %v:%v }", doc.Key, toLegacyString(doc.Value)))
				}
			} else {
				arr = append(arr, fmt.Sprintf("%v", alist))
			}
			arrs = append(arrs, strings.Join(arr, ", "))
		}
		return "[" + strings.Join(arrs, ", ") + "]"
	} else if list, ok := o.(bson.D); ok {
		arr := []string{}
		for _, doc := range list {
			if _, ok := doc.Value.(bson.D); ok {
				b, _ := bson.MarshalExtJSON(doc.Value, false, false)
				arr = append(arr, fmt.Sprintf("%v: %v", doc.Key, string(b)))
			} else {
				arr = append(arr, fmt.Sprintf("%v:%v", doc.Key, toLegacyString(doc.Value)))
			}
		}
		return " { " + strings.Join(arr, ", ") + " }"
	} else if elem, ok := o.(bson.E); ok {
		return fmt.Sprintf(" { %v:%v } ", elem.Key, toLegacyString(elem.Value))
	} else {
		if _, ok := o.(string); ok {
			return fmt.Sprintf(` %v`, o)
		} else if _, ok := o.(bool); ok {
			return fmt.Sprintf(` %v`, o)
		} else {
			return o
		}
	}
}
