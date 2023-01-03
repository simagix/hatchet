// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"html/template"
	"regexp"
)

// GetLogsTemplate returns HTML
func GetLogsTemplate() (*template.Template, error) {
	html := headers + menuHTML + getLogsTable() + "</body>"
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
	html := headers + menuHTML + getLegacyLogTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
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
<div><p/>{{.Summary}}</div>
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
<div align='center' width='100%'>
	<select id='component' class="btn2" style="float: left;">
		<option value=''>select a component</option>
		<option value='ACCESS'>ACCESS</option>
		<option value='ASIO'>ASIO</option>
		<option value='COMMAND'>COMMAND</option>
		<option value='CONNPOOL'>CONNPOOL</option>
		<option value='CONTROL'>CONTROL</option>
		<option value='ELECTION'>ELECTION</option>
		<option value='FTDC'>FTDC</option>
		<option value='INDEX'>INDEX</option>
		<option value='INITSYNC'>INITSYNC</option>
		<option value='NETWORK'>NETWORK</option>
		<option value='QUERY'>QUERY</option>
		<option value='RECOVERY'>RECOVERY</option>
		<option value='REPL'>REPL</option>
		<option value='SHARDING'>SHARDING</option>
		<option value='STORAGE'>STORAGE</option>
		<option value='WRITE'>WRITE</option>
	</select>

	<select id='severity' class="btn2" style="float: center;">
		<option value=''>select a severity</option>
		<option value='F'>Fatal</option>
		<option value='E'>Error</option>
		<option value='W'>Warn</option>
		<option value='I'>Info</option>
	</select>

	<button id="find" onClick="findLogs()" class="btn" style="float: right;">Find</button>
	<input id='context' type='text' class="btn2" style="float: right;" size='25'/>
	<label class="btn2" style="float: right;">context: </label>
</div>
<p/>
<div align='center'>
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
</div>

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
