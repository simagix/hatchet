// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"errors"
	"fmt"
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

// Pre-compiled regex patterns for better performance
var (
	reMatchSort = regexp.MustCompile(`^{("\$match"|"\$sort"):(\S+)}$`)
	reFacet     = regexp.MustCompile(`^{("(\$facet")):\S+}$`)
	reOid       = regexp.MustCompile(`{"\$oid":1}`)
	reIn        = regexp.MustCompile(`("\$n?in"):\[[^\]]+\]`)
	reKey       = regexp.MustCompile(`"(\$?\w+)":`)
	reRegex     = regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
	mapWalker   = gox.NewMapWalker(cb) // Reusable walker instance
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
	// Parse attributes directly from bson.D instead of Marshal/Unmarshal
	parseAttributes(doc)
	stat.TotalMilli = doc.Attributes.Milli
	stat.Namespace = doc.Attributes.NS
	if stat.Namespace == "" {
		return stat, errors.New("no namespace found")
	} else if strings.HasSuffix(stat.Namespace, ".$cmd") {
		// Extract actual namespace from command for .$cmd namespaces
		if commands, ok := BsonD2M(doc.Attr)["command"].(bson.D); ok {
			if len(commands) > 0 {
				elem := commands[0]
				if ns, ok := elem.Value.(string); ok {
					if strings.Contains(ns, ".") {
						doc.Attributes.NS = ns
					} else {
						doc.Attributes.NS = strings.ReplaceAll(stat.Namespace, "$cmd", ns)
					}
				}
			}
		}
		// Update stat.Namespace with corrected value
		stat.Namespace = doc.Attributes.NS
		if doc.Attributes.ErrMsg != "" {
			stat.Index = "ErrMsg: " + doc.Attributes.ErrMsg
			return stat, errors.New("unrecognized collection, ignored")
		}
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
	if stat.Op == "command" || stat.Op == "none" || stat.Op == "" {
		stat.Op = getOp(command)
	}
	var isGetMore bool
	if stat.Op == cmdGetMore {
		isGetMore = true
		if doc.Attributes.OriginatingCommand != nil {
			command = doc.Attributes.OriginatingCommand
			stat.Op = getOp(command)
		}
	}
	if stat.Op == cmdInsert || stat.Op == cmdCollstats {
		stat.QueryPattern = ""
	} else if stat.Op == cmdDistinct {
		// Show which field is being distincted
		if key, ok := command["key"].(string); ok {
			stat.QueryPattern = fmt.Sprintf("{ key:%q }", key)
		}
	} else if stat.Op == cmdCount {
		// Extract query from count command
		if query := command["query"]; query != nil {
			if qmap := toMap(query); qmap != nil {
				walked := mapWalker.Walk(qmap)
				if buf, err := json.Marshal(walked); err == nil {
					stat.QueryPattern = string(buf)
				}
			}
		}
		if stat.QueryPattern == "" {
			stat.QueryPattern = "{}"
		}
	} else if stat.Op == cmdCreateIndexes {
		// Extract the index key from createIndexes command and show in Index field
		stat.QueryPattern = ""
		if indexes := command["indexes"]; indexes != nil {
			var indexArr []interface{}
			switch idx := indexes.(type) {
			case bson.A:
				indexArr = idx
			case []interface{}:
				indexArr = idx
			}
			if len(indexArr) > 0 {
				if indexDoc := toMap(indexArr[0]); indexDoc != nil {
					if key := indexDoc["key"]; key != nil {
						if keyMap := toMap(key); keyMap != nil {
							if buf, err := json.Marshal(keyMap); err == nil {
								stat.Index = string(buf)
								stat.Index = reKey.ReplaceAllString(stat.Index, ` $1:`)
								stat.Index = strings.ReplaceAll(stat.Index, "}", " }")
							}
						}
					}
				}
			}
		}
	} else if stat.Op == cmdFind || stat.Op == cmdUpdate || stat.Op == cmdRemove || stat.Op == cmdDelete || stat.Op == cmdFindAndModify {
		// Extract query/filter for find, update, delete, remove, findAndModify
		var query interface{}
		if command["q"] != nil {
			query = command["q"]
		} else if command["query"] != nil {
			query = command["query"]
		} else if command["filter"] != nil {
			query = command["filter"]
		}

		if query != nil {
			qmap := toMap(query)
			if qmap != nil {
				walked := mapWalker.Walk(qmap)
				if buf, err := json.Marshal(walked); err == nil {
					stat.QueryPattern = string(buf)
				} else {
					stat.QueryPattern = "{}"
				}
			} else {
				stat.QueryPattern = "{}"
			}
		} else {
			stat.QueryPattern = "{}"
		}
	} else if stat.Op == cmdAggregate {
		var pipeline []interface{}
		switch p := command["pipeline"].(type) {
		case bson.A:
			pipeline = p
		case []interface{}:
			pipeline = p
		default:
			return stat, errors.New("pipeline not found")
		}
		if len(pipeline) == 0 {
			return stat, errors.New("pipeline not found")
		}
		// Find the first $match stage
		var matchStage interface{}
		for _, v := range pipeline {
			stageMap := toMap(v)
			if stageMap != nil {
				if _, ok := stageMap["$match"]; ok {
					matchStage = v
					break
				}
			}
		}
		// If $match found, extract its filter pattern
		if matchStage != nil {
			fmap := toMap(matchStage)
			if fmap != nil && !isRegex(fmap) {
				walked := mapWalker.Walk(fmap)
				if buf, err := json.Marshal(walked); err == nil {
					stat.QueryPattern = string(buf)
				} else {
					stat.QueryPattern = "{}"
				}
			} else if fmap != nil {
				buf, _ := json.Marshal(fmap)
				stat.QueryPattern = reRegex.ReplaceAllString(string(buf), "{$1:/$3.../$2}")
			}
		} else {
			// No $match - show full pipeline with normalized values
			stat.QueryPattern = getPipelinePattern(pipeline)
		}
	} else {
		var fmap map[string]interface{}
		if command["filter"] != nil {
			fmap = toMap(command["filter"])
		} else if command["query"] != nil {
			fmap = toMap(command["query"])
		} else if command["q"] != nil {
			fmap = toMap(command["q"])
		} else {
			stat.QueryPattern = "{}"
		}
		if fmap != nil && !isRegex(fmap) {
			walked := mapWalker.Walk(fmap)
			var data []byte
			if data, err = json.Marshal(walked); err != nil {
				return stat, err
			}
			stat.QueryPattern = string(data)
			if stat.QueryPattern == `{"":null}` {
				stat.QueryPattern = "{}"
			}
		} else if fmap != nil {
			buf, _ := json.Marshal(fmap)
			stat.QueryPattern = reRegex.ReplaceAllString(string(buf), "{$1:/$3.../$2}")
		}
	}
	if stat.Op == "" {
		return stat, nil
	}
	// Use pre-compiled regex patterns for better performance
	stat.QueryPattern = reMatchSort.ReplaceAllString(stat.QueryPattern, `$2`)
	stat.QueryPattern = reFacet.ReplaceAllString(stat.QueryPattern, `{$1:...}`)
	stat.QueryPattern = reOid.ReplaceAllString(stat.QueryPattern, `1`)
	stat.QueryPattern = reIn.ReplaceAllString(stat.QueryPattern, `$1:[...]`)
	stat.QueryPattern = reKey.ReplaceAllString(stat.QueryPattern, ` $1:`)
	stat.QueryPattern = strings.ReplaceAll(stat.QueryPattern, "}", " }")
	if isGetMore {
		stat.Op = cmdGetMore
	}
	return stat, nil
}

// parseAttributes extracts attributes directly from bson.D without Marshal/Unmarshal
func parseAttributes(doc *Logv2Info) {
	for _, elem := range doc.Attr {
		switch elem.Key {
		case "appName":
			if v, ok := elem.Value.(string); ok {
				doc.Attributes.AppName = v
			}
		case "command":
			if v, ok := elem.Value.(bson.D); ok {
				doc.Attributes.Command = BsonD2M(v)
			} else if v, ok := elem.Value.(map[string]interface{}); ok {
				doc.Attributes.Command = v
			}
		case "errMsg":
			if v, ok := elem.Value.(string); ok {
				doc.Attributes.ErrMsg = v
			}
		case "durationMillis":
			doc.Attributes.Milli = toInt(elem.Value)
		case "ns":
			if v, ok := elem.Value.(string); ok {
				doc.Attributes.NS = v
			}
		case "originatingCommand":
			if v, ok := elem.Value.(bson.D); ok {
				doc.Attributes.OriginatingCommand = BsonD2M(v)
			} else if v, ok := elem.Value.(map[string]interface{}); ok {
				doc.Attributes.OriginatingCommand = v
			}
		case "planSummary":
			if v, ok := elem.Value.(string); ok {
				doc.Attributes.PlanSummary = v
			}
		case "reslen":
			doc.Attributes.Reslen = toInt(elem.Value)
		case "type":
			if v, ok := elem.Value.(string); ok {
				doc.Attributes.Type = v
			}
		}
	}
}

// toInt converts various numeric types to int efficiently
func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	}
	return 0
}

// toMap converts bson.D or map[string]interface{} to map[string]interface{}
func toMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}
	switch m := v.(type) {
	case map[string]interface{}:
		return m
	case bson.D:
		return BsonD2M(m)
	case bson.M:
		return m
	}
	return nil
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
		cmdFind, cmdFindAndModify, cmdGetMore, cmdInsert, cmdRemove, cmdUpdate}
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

// getPipelinePattern returns the full pipeline with values normalized to 1
func getPipelinePattern(pipeline []interface{}) string {
	// Walk the entire pipeline to normalize values
	var normalizedPipeline []interface{}
	for _, stage := range pipeline {
		stageMap := toMap(stage)
		if stageMap != nil {
			walked := mapWalker.Walk(stageMap)
			normalizedPipeline = append(normalizedPipeline, walked)
		}
	}
	if len(normalizedPipeline) == 0 {
		return "{}"
	}
	buf, err := json.Marshal(normalizedPipeline)
	if err != nil {
		return "{}"
	}
	return string(buf)
}
