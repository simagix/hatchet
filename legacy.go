// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AddLegacyString converts log to legacy format
func AddLegacyString(doc *Logv2Info) error {
	var err error
	var arr []string

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

	if doc.Component == "CONTROL" && doc.Attr["host"] != nil {
		arr = append(arr, fmt.Sprintf("pid=%v port=%v %v host=%v", doc.Attr["pid"], doc.Attr["port"], doc.Attr["architecture"], doc.Attr["host"]))
	} else if doc.Component == "ACCESS" {
		var milli interface{}
		for k, v := range doc.Attr {
			if k == "authenticationDatabase" {
				arr = append(arr, fmt.Sprintf("on %v", v))
			} else if k == "principalName" {
				arr = append(arr, fmt.Sprintf("as principal %v", v))
			} else if k == "remote" {
				arr = append(arr, fmt.Sprintf("from client %v", v))
			} else if k == "durationMillis" {
				milli = v
			} else {
				b, _ := json.Marshal(v)
				arr = append(arr, fmt.Sprintf("%v:%v", k, string(b)))
			}
		}
		if milli != nil {
			arr = append(arr, fmt.Sprintf("%vms", milli))
		}
	} else if doc.Component == "NETWORK" {
		for k, v := range doc.Attr {
			if k == "remote" {
				if doc.Msg == "Connection ended" {
					arr = append(arr, fmt.Sprintf("%v", v))
				} else {
					arr = append(arr, fmt.Sprintf("from %v", v))
				}
			} else if k == "client" {
				arr = append(arr, fmt.Sprintf("%v:", v))
			} else if k == "connectionId" && doc.Msg != "Connection ended" {
				arr = append(arr, fmt.Sprintf("#%v", v))
			} else if k == "connectionCount" {
				arr = append(arr, fmt.Sprintf("(%v connections now open)", v))
			} else {
				b, _ := json.Marshal(v)
				arr = append(arr, string(b))
			}
		}
	} else if doc.Component == "COMMAND" || doc.Component == "WRITE" || doc.Component == "QUERY" || doc.Component == "TXT" {
		var milli interface{}
		for k, v := range doc.Attr {
			if k == "type" || k == "ns" {
				b, _ := json.Marshal(v)
				arr = append(arr, string(b))
			} else if k == "command" {
				b, _ := json.Marshal(v)
				attrs := string(b)
				length := len(attrs)
				arr = append(arr, attrs[1:length-1])
			} else if k == "durationMillis" {
				milli = v
			} else {
				b, _ := json.Marshal(v)
				arr = append(arr, fmt.Sprintf("%v:%v", k, string(b)))
			}
		}
		if milli != nil {
			arr = append(arr, fmt.Sprintf("%vms", milli))
		}
	}

	if len(arr) == 0 {
		return nil
	}
	doc.Message = strings.Join(arr, " ")
	return err
}
