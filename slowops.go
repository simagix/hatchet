// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	cmdCollstats     = "collstats"
	cmdCreateIndexes = "createIndexes"

	cmdInsert   = "insert"
	cmdDistinct = "distinct"

	cmdAggregate     = "aggregate"
	cmdCount         = "count"
	cmdDelete        = "delete"
	cmdFind          = "find"
	cmdFindAndModify = "findandmodify"
	cmdGetMore       = "getMore"
	cmdRemove        = "remove"
	cmdUpdate        = "update"
)

// AnalyzeLog analyzes slow op log
func AnalyzeLog(str string) (*OpStat, error) {
	doc := Logv2Info{}
	if err := bson.UnmarshalExtJSON([]byte(str), false, &doc); err != nil {
		return nil, err
	}
	return AnalyzeSlowOp(&doc)
}

// AnalyzeSlowOp analyzes slow ops
func AnalyzeSlowOp(doc *Logv2Info) (*OpStat, error) {
	var err error
	stat := &OpStat{}

	c := doc.Component
	if c != "COMMAND" && c != "QUERY" && c != "WRITE" {
		return stat, errors.New("unsupported command")
	}
	b, _ := bson.Marshal(doc.Attr)
	bson.Unmarshal(b, &doc.Attributes)
	stat.TotalMilli = doc.Attributes.Milli
	stat.Namespace = doc.Attributes.NS
	if stat.Namespace == "" {
		return stat, errors.New("no namespace found")
	} else if strings.HasSuffix(stat.Namespace, ".$cmd") {
		if commands, ok := doc.Attr.Map()["command"].(bson.D); ok {
			if len(commands) > 0 {
				elem := commands[0]
				stat.Op = elem.Key
				if doc.Attributes.NS, ok = elem.Value.(string); !ok {
					doc.Attributes.NS = stat.Namespace
				}
				if !strings.Contains(doc.Attributes.NS, ".") {
					doc.Attributes.NS = strings.ReplaceAll(stat.Namespace, "$cmd", doc.Attributes.NS)
				}
			}
		}
		if doc.Attributes.ErrMsg != "" {
			stat.Index = "ErrMsg: " + doc.Attributes.ErrMsg
		}
		return stat, errors.New("system command")
	}
	if doc.Attributes.PlanSummary != "" { // not insert
		plan := doc.Attributes.PlanSummary
		if strings.HasPrefix(plan, "IXSCAN") {
			stat.Index = plan[len("IXSCAN")+1:]
			stat.Index = strings.ReplaceAll(stat.Index, ": ", ":")
		} else {
			stat.Index = plan
		}
	} else if doc.Attributes.ErrMsg != "" {
		stat.Index = "ErrMsg: " + doc.Attributes.ErrMsg
	}
	stat.Reslen = doc.Attributes.Reslen
	if doc.Attributes.Command == nil {
		return stat, errors.New("no command found")
	}
	command := doc.Attributes.Command
	stat.Op = doc.Attributes.Type
	if stat.Op == "command" || stat.Op == "none" {
		stat.Op = getOp(command)
	}
	var isGetMore bool
	if stat.Op == cmdGetMore {
		isGetMore = true
		command = doc.Attributes.OriginatingCommand
		stat.Op = getOp(command)
	}
	if stat.Op == cmdInsert || stat.Op == cmdDistinct ||
		stat.Op == cmdCreateIndexes || stat.Op == cmdCollstats {
		stat.QueryPattern = ""
	} else if stat.QueryPattern == "" &&
		(stat.Op == cmdFind || stat.Op == cmdUpdate || stat.Op == cmdRemove || stat.Op == cmdDelete) {
		var query interface{}
		if command["q"] != nil {
			query = command["q"]
		} else if command["query"] != nil {
			query = command["query"]
		} else if command["filter"] != nil {
			query = command["filter"]
		}

		if query != nil {
			walker := gox.NewMapWalker(cb)
			doc := walker.Walk(query.(map[string]interface{}))
			if buf, err := json.Marshal(doc); err == nil {
				stat.QueryPattern = string(buf)
			} else {
				stat.QueryPattern = "{}"
			}
		} else {
			stat.QueryPattern = "{}"
		}
	} else if stat.Op == cmdAggregate {
		pipeline, ok := command["pipeline"].(bson.A)
		if !ok || len(pipeline) == 0 {
			return stat, errors.New("pipeline not found")
		}
		var stage interface{}
		for _, v := range pipeline {
			stage = v
			break
		}
		fmap := stage.(map[string]interface{})
		if !isRegex(fmap) {
			walker := gox.NewMapWalker(cb)
			doc := walker.Walk(fmap)
			if buf, err := json.Marshal(doc); err == nil {
				stat.QueryPattern = string(buf)
			} else {
				stat.QueryPattern = "{}"
			}
			if strings.Contains(stat.QueryPattern, "$changeStream") {
				if len(pipeline) > 1 {
					buf, _ := json.Marshal(pipeline[1])
					stat.QueryPattern = string(buf)
				} else {
					stat.QueryPattern = "{}"
				}
			} else if !strings.Contains(stat.QueryPattern, "$match") && !strings.Contains(stat.QueryPattern, "$sort") &&
				!strings.Contains(stat.QueryPattern, "$facet") && !strings.Contains(stat.QueryPattern, "$indexStats") {
				stat.QueryPattern = "{}"
			}
		} else {
			buf, _ := json.Marshal(fmap)
			str := string(buf)
			re := regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
			stat.QueryPattern = re.ReplaceAllString(str, "{$1:/$3.../$2}")
		}
	} else {
		var fmap map[string]interface{}
		if command["filter"] != nil {
			fmap = command["filter"].(map[string]interface{})
		} else if command["query"] != nil {
			fmap = command["query"].(map[string]interface{})
		} else if command["q"] != nil {
			fmap = command["q"].(map[string]interface{})
		} else {
			stat.QueryPattern = "{}"
		}
		if !isRegex(fmap) {
			walker := gox.NewMapWalker(cb)
			doc := walker.Walk(fmap)
			var data []byte
			if data, err = json.Marshal(doc); err != nil {
				return stat, err
			}
			stat.QueryPattern = string(data)
			if stat.QueryPattern == `{"":null}` {
				stat.QueryPattern = "{}"
			}
		} else {
			buf, _ := json.Marshal(fmap)
			str := string(buf)
			re := regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
			stat.QueryPattern = re.ReplaceAllString(str, "{$1:/$3.../$2}")
		}
	}
	if stat.Op == "" {
		return stat, nil
	}
	re := regexp.MustCompile(`^{("\$match"|"\$sort"):(\S+)}$`)
	stat.QueryPattern = re.ReplaceAllString(stat.QueryPattern, `$2`)
	re = regexp.MustCompile(`^{("(\$facet")):\S+}$`)
	stat.QueryPattern = re.ReplaceAllString(stat.QueryPattern, `{$1:...}`)
	re = regexp.MustCompile(`{"\$oid":1}`)
	stat.QueryPattern = re.ReplaceAllString(stat.QueryPattern, `1`)
	re = regexp.MustCompile(`("\$n?in"):\[\S+(,\s?\S+)*\]`)
	stat.QueryPattern = re.ReplaceAllString(stat.QueryPattern, `$1:[...]`)
	re = regexp.MustCompile(`"(\$?\w+)":`)
	stat.QueryPattern = re.ReplaceAllString(stat.QueryPattern, ` $1:`)
	stat.QueryPattern = strings.ReplaceAll(stat.QueryPattern, "}", " }")
	if isGetMore {
		stat.Op = cmdGetMore
	}
	return stat, nil
}

func isRegex(doc map[string]interface{}) bool {
	if buf, err := json.Marshal(doc); err != nil {
		return false
	} else if strings.Contains(string(buf), "$regularExpression") {
		return true
	}
	return false
}

func getOp(command map[string]interface{}) string {
	ops := []string{cmdAggregate, cmdCollstats, cmdCount, cmdCreateIndexes, cmdDelete, cmdDistinct,
		cmdFind, cmdFindAndModify, cmdGetMore, cmdInsert, cmdUpdate}
	for _, v := range ops {
		if command[v] != nil {
			return v
		}
	}
	return ""
}

func cb(value interface{}) interface{} {
	return 1
}
