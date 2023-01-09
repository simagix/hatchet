// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

// GetLogsTemplate returns HTML
func GetLogsTemplate() (*template.Template, error) {
	html := getContentHTML("", "") + getLogsTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"highligtLog": func(log string) template.HTML {
			return template.HTML(highlightLog(log))
		}}).Parse(html)
}

// GetLegacyLogTemplate returns HTML
func GetLegacyLogTemplate() (*template.Template, error) {
	html := getContentHTML("", "") + getLegacyLogTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"getComponentOptions": func(item string) template.HTML {
			arr := []string{}
			comps := []string{"ACCESS", "ASIO", "COMMAND", "CONNPOOL", "CONTROL", "ELECTION", "FTDC", "INDEX", "INITSYNC", "NETWORK",
				"QUERY", "RECOVERY", "REPL", "SHARDING", "STORAGE", "WRITE"}
			for _, v := range comps {
				selected := ""
				if v == item {
					selected = "SELECTED"
				}
				arr = append(arr, fmt.Sprintf("<option value='%v' %v>%v</option>", v, selected, v))
			}
			return template.HTML(strings.Join(arr, "\n"))
		},
		"getSeverityOptions": func(item string) template.HTML {
			arr := []string{}
			servs := [][2]string{{"FATAL", "F"}, {"ERROR", "E"}, {"WARN", "W"}, {"INFO", "I"}, {"DEBUG", "D"}, {"DEBUG2", "D2"}}
			for _, v := range servs {
				selected := ""
				if v[1] == item {
					selected = "SELECTED"
				}
				arr = append(arr, fmt.Sprintf("<option value='%v' %v>%v</option>", v[1], selected, v[0]))
			}
			return template.HTML(strings.Join(arr, "\n"))
		},
		"highligtLog": func(log string) template.HTML {
			return template.HTML(highlightLog(log))
		}}).Parse(html)
}

func highlightLog(log string) string {
	re := regexp.MustCompile(`("?planSummary"?:\s?"?\w+"?)`)
	log = re.ReplaceAllString(log, "<mark>$1</mark>")
	re = regexp.MustCompile(`((\d+ms$))`)
	log = re.ReplaceAllString(log, "<mark>$1</mark>")
	re = regexp.MustCompile(`(("?(keysExamined|keysInserted|docsExamined|nreturned|nMatched|nModified|ndeleted|ninserted|reslen)"?:)\d+)`)
	return re.ReplaceAllString(log, "<mark>$1</mark>")
}

func getLogsTable() string {
	template := ` 
<p/>
<div align='center'>
	<table width='100%'>
		<tr>
			<th>#</th>
			<th>log in legacy format</th>
		</tr>
{{range $n, $value := .Logs}}
		<tr>
			<td align='right'>{{ add $n 1 }}</td>
			<td>{{ highligtLog $value }}</td>
		</tr>
{{end}}
	</table>
</div>
`
	return template
}

func getLegacyLogTable() string {
	template := `
<br/>
<div style="float: left;">
	<label>component: </label>
	<select id='component'>
		<option value=''>select a component</option>
		{{getComponentOptions .Component}}
	</select>
</div>

<div style="float: left; padding: 0px 0px 0px 20px;">
	<label>&nbsp;severity: </label>
	<select id='severity'>
		<option value=''>select a severity</option>
		{{getSeverityOptions .Severity}}
	</select>
</div>

<div style="float: left; padding: 0px 0px 0px 20px;">
	<label>&nbsp;context: </label>
	<input id='context' type='text' value='{{.Context}}' size='25'/>
	<button id="find" onClick="findLogs()" class="button" style="float: right;">Find</button>
</div>

<p/>
<div>
{{ if gt .LogLength 0 }}
	<table width='100%'>
		<tr>
			<th>#</th>
			<th>date</th>
			<th>S</th>
			<th>component</th>
			<th>context</th>
			<th>message</th>
		</tr>
	{{range $n, $value := .Logs}}
		<tr>
			<td align='right'>{{ add $n 1 }}</td>
			<td>{{ $value.Timestamp }}</td>
			<td>{{ $value.Severity }}</td>
			<td>{{ $value.Component }}</td>
			<td>{{ $value.Context }}</td>
			<td>{{ highligtLog $value.Message }}</td>
		</tr>
	{{end}}
	</table>
{{end}}
</div>
<p/>
<script>
	var input = document.getElementById("context");
	input.addEventListener("keypress", function(event) {
		if (event.key === "Enter") {
			event.preventDefault();
			document.getElementById("find").click();
		}
	});

	function findLogs() {
		var sel = document.getElementById('component')
		var component = sel.options[sel.selectedIndex].value;
		sel = document.getElementById('severity')
		var severity = sel.options[sel.selectedIndex].value;
		var context = document.getElementById('context').value
		window.location.href = '/tables/{{.Table}}/logs?component='+component+'&severity='+severity+'&context='+context;
	}
</script>
`
	return template
}
