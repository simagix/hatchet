// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/simagix/gox"
)

var ops = []string{cmdAggregate, cmdCount, cmdDelete, cmdDistinct, cmdFind,
	cmdFindAndModify, cmdGetMore, cmdInsert, cmdUpdate}

const cmdAggregate = "aggregate"
const cmdCount = "count"
const cmdCreateIndexes = "createIndexes"
const cmdDelete = "delete"
const cmdDistinct = "distinct"
const cmdFind = "find"
const cmdFindAndModify = "findandmodify"
const cmdGetMore = "getMore"
const cmdInsert = "insert"
const cmdRemove = "remove"
const cmdUpdate = "update"

// AnalyzeSlowQuery analyzes slow queries
func AnalyzeSlowQuery(doc *Logv2Info) (LogStats, error) {
	var err error
	var stat = LogStats{}

	c := doc.Component
	if c != "COMMAND" && c != "QUERY" && c != "WRITE" {
		return stat, errors.New("unsupported command")
	}
	stat.milli = doc.Attributes.Milli
	stat.ns = doc.Attributes.NS
	if stat.ns == "" {
		return stat, errors.New("no namespace found")
	} else if strings.HasPrefix(stat.ns, "admin.") || strings.HasPrefix(stat.ns, "config.") || strings.HasPrefix(stat.ns, "local.") {
		stat.op = DOLLAR_CMD
		return stat, errors.New("system database")
	} else if strings.HasSuffix(stat.ns, ".$cmd") {
		stat.op = DOLLAR_CMD
		return stat, errors.New("system command")
	}

	if doc.Attributes.PlanSummary != "" { // not insert
		plan := doc.Attributes.PlanSummary
		if plan == COLLSCAN {
			stat.index = ""
		} else if strings.HasPrefix(plan, "IXSCAN") {
			stat.index = plan[len("IXSCAN")+1:]
		} else {
			stat.index = plan
		}
	}
	stat.reslen = doc.Attributes.Reslen
	if doc.Attributes.Command == nil {
		return stat, errors.New("no command found")
	}
	command := doc.Attributes.Command
	stat.op = doc.Attributes.Type
	if stat.op == "command" || stat.op == "none" {
		stat.op = getOp(command)
	}
	var isGetMore bool
	if stat.op == cmdGetMore {
		isGetMore = true
		command = doc.Attributes.OriginatingCommand
		stat.op = getOp(command)
	}
	if stat.op == cmdInsert || stat.op == cmdCreateIndexes {
		stat.filter = "N/A"
	} else if (stat.op == cmdUpdate || stat.op == cmdRemove || stat.op == cmdDelete) && stat.filter == "" {
		var query interface{}
		if command["q"] != nil {
			query = command["q"]
		} else if command["query"] != nil {
			query = command["query"]
		}

		if query != nil {
			walker := gox.NewMapWalker(cb)
			doc := walker.Walk(query.(map[string]interface{}))
			if buf, err := json.Marshal(doc); err == nil {
				stat.filter = string(buf)
			} else {
				stat.filter = "{}"
			}
		} else {
			return stat, errors.New("no filter found")
		}
	} else if stat.op == cmdAggregate {
		pipeline, ok := command["pipeline"].([]interface{})
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
				stat.filter = string(buf)
			} else {
				stat.filter = "{}"
			}
			if !strings.Contains(stat.filter, "$match") && !strings.Contains(stat.filter, "$sort") &&
				!strings.Contains(stat.filter, "$facet") && !strings.Contains(stat.filter, "$indexStats") {
				stat.filter = "{}"
			}
		} else {
			buf, _ := json.Marshal(fmap)
			str := string(buf)
			re := regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
			stat.filter = re.ReplaceAllString(str, "{$1:/$3.../$2}")
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
			return stat, errors.New("no filter found")
		}
		if !isRegex(fmap) {
			walker := gox.NewMapWalker(cb)
			doc := walker.Walk(fmap)
			var data []byte
			if data, err = json.Marshal(doc); err != nil {
				return stat, err
			}
			stat.filter = string(data)
			if stat.filter == `{"":null}` {
				stat.filter = "{}"
			}
		} else {
			buf, _ := json.Marshal(fmap)
			str := string(buf)
			re := regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
			stat.filter = re.ReplaceAllString(str, "{$1:/$3.../$2}")
		}
	}
	if stat.op == "" {
		return stat, nil
	}
	re := regexp.MustCompile(`\[1(,1)*\]`)
	stat.filter = re.ReplaceAllString(stat.filter, `[...]`)
	re = regexp.MustCompile(`\[{\S+}(,{\S+})*\]`) // matches repeated doc {"base64":1,"subType":1}}
	stat.filter = re.ReplaceAllString(stat.filter, `[...]`)
	re = regexp.MustCompile(`^{("\$match"|"\$sort"):(\S+)}$`)
	stat.filter = re.ReplaceAllString(stat.filter, `$2`)
	re = regexp.MustCompile(`^{("(\$facet")):\S+}$`)
	stat.filter = re.ReplaceAllString(stat.filter, `{$1:...}`)
	re = regexp.MustCompile(`{"\$oid":1}`)
	stat.filter = re.ReplaceAllString(stat.filter, `1`)
	if isGetMore {
		stat.op = cmdGetMore
	}
	stat.date = doc.Timestamp["$date"]
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
