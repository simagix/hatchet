// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"html/template"
	"regexp"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetStatsTemplate returns HTML
func GetStatsTemplate() (*template.Template, error) {
	html := headers + menuHTML + getStatsTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"numPrinter": func(n interface{}) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", ToInt(n))
		}}).Parse(html)
}

// GetLogsTemplate returns HTML
func GetLogsTemplate() (*template.Template, error) {
	html := headers + menuHTML + getLogsTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"highligtLog": func(log string) template.HTML {
			re := regexp.MustCompile(`("?planSummary"?:\s?"?\w+"?)`)
			log = re.ReplaceAllString(log, "<mark>$1</mark>")
			re = regexp.MustCompile(`((\d+ms$))`)
			log = re.ReplaceAllString(log, "<mark>$1</mark>")
			re = regexp.MustCompile(`(("?(keysExamined|keysInserted|docsExamined|nreturned|nMatched|nModified|ndeleted|ninserted|reslen)"?:)\d+)`)
			return template.HTML(re.ReplaceAllString(log, "<mark>$1</mark>"))
		}}).Parse(html)
}

// GetLegacyLogTemplate returns HTML
func GetLegacyLogTemplate() (*template.Template, error) {
	html := headers + menuHTML + getLegacyLogTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		}}).Parse(html)
}

// GetTablesTemplate returns HTML
func GetTablesTemplate() (*template.Template, error) {
	html := headers + getMainPage() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		}}).Parse(html)
}

const headers = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Ken Chen's Hatchet</title>
	<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
	<meta http-equiv="Pragma" content="no-cache" />
	<meta http-equiv="Expires" content="0" />

  <script src="https://www.gstatic.com/charts/loader.js"></script>
  <link href="/favicon.ico" rel="icon" type="image/x-icon" />
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">
  <style>
  	body {
      font-family: Helvetica, Arial, sans-serif;
      margin-top: 10px;
      margin-bottom: 10px;
      margin-right: 10px;
      margin-left: 10px;
    }
    table {
      font-family: Consolas, monaco, monospace;
      border-collapse:collapse;
      min-width:600px;
    }
    caption {
      caption-side:top;
      font-weight:bold;
      font-style:italic;
      margin:2px;
    }
    table, th, td {
      border: 1px solid gray;
      vertical-align: top;
    }
    th, td {
      padding:2px;
      vertical-align: top;
      font-size: 11px;
    }
    th {
      background-color: #ddd;
      font-weight:bold;
    }
    tr:nth-child(even) {background-color: #f2f2f2;}
    tr:nth-child(odd) {background-color: #fff;}
    .api {
      font-family: Consolas, monaco, monospace;
    }
    .btn {
      background-color: #fff;
      border: none;
      outline:none;
      color: #4285F4;
      padding: 5px 10px;
      cursor: pointer;
      font-size: 16px;
    }
    .btn:hover {
      color: #DB4437;
    }
    .button {
      background-color: #fff;
      border: none;
      outline:none;
      color: #000;
      padding: 5px 5px;
      cursor: pointer;
      font-size: 16px;
    }
    h1 {
      font-family: "Trebuchet MS";
      font-size: 1.7em;
      font-weight: bold;
    }
    h2 {
      font-family: "Trebuchet MS";
      font-size: 1.5em;
      font-weight: bold;
    }
    h3 {
      font-family: "Trebuchet MS";
      font-size: 1.25em;
      font-weight: bold;
    }
    h4 {
      font-family: "Trebuchet MS";
      font-size: 1em;
      font-weight: bold;
    }
  </style>
</head>

<body>
`

const menuHTML = `
<script>
	function gotoChart() {
		var sel = document.getElementById('nextChart')
		var value = sel.options[sel.selectedIndex].value;
		if(value == "") {
			return;
		}
		window.location.href = value
	}
</script>
<div align='center'>
	<select id='nextChart' class='btn' style="float: right;" onchange='gotoChart()'>
		<option value=''>select a chart</option>
		<option value='/tables/{{.Table}}/charts/slowops'>Ops Stats</option>
		<option value='/tables/{{.Table}}/charts/accepted_conns'>Accepted Connections</option>
		<option value='/tables/{{.Table}}/charts/connections?type=time'>Connections by Time</option>
		<option value='/tables/{{.Table}}/charts/connections?type=total'>Connections by Total</option>
	</select>
	<button id="title" onClick="javascript:location.href='/'; return false;"
		class="btn" style="float: center;"><i class="fa fa-leaf"></i> Hatchet</button>
	<button id="stats" onClick="javascript:location.href='/tables/{{.Table}}/stats/slowops'; return false;"
		class="btn" style="float: left;"><i class="fa fa-info"></i> Stats</button>
	<button id="chart" class="btn" style="float: right;"><i class="fa fa-bar-chart"></i></button>
	<button id="logs" onClick="javascript:location.href='/tables/{{.Table}}/logs/slowops'; return false;"
		class="btn" style="float: left;"><i class="fa fa-list"></i> Top N</button>
</div>
`

func getStatsTable() string {
	template := `
<div align='center'>
	<table width='100%'>
		<tr>
			<th>#</th>
			<th>command</th>
			<th>namespace</th>
			<th>count</th>
			<th>avg ms</th>
			<th>max ms</th>
			<th>total ms</th>
			<th>total reslen</th>
			<th>index</th>
			<th>query pattern</th>
		</tr>
{{range $n, $value := .Ops}}
		<tr>
			<td align='right'>{{ add $n 1 }}</td>
			<td>{{ $value.Op }}</td>
			<td>{{ $value.Namespace }}</td>
			<td align='right'>{{ numPrinter $value.Count }}</td>
			<td align='right'>{{ numPrinter $value.AvgMilli }}</td>
			<td align='right'>{{ numPrinter $value.MaxMilli }}</td>
			<td align='right'>{{ numPrinter $value.TotalMilli }}</td>
			<td align='right'>{{ numPrinter $value.Reslen }}</td>
		{{ if ( eq $value.Index "COLLSCAN" ) }}
			<td><span style="color:red;">{{ $value.Index }}</span></td>
		{{else}}
			<td>{{ $value.Index }}</td>
		{{end}}
			<td>{{ $value.QueryPattern }}</td>
		</tr>
{{end}}
	</table>
</div>
`
	return template
}

func getLogsTable() string {
	template := `
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
			<td>{{ $value.Message }}</td>
		</tr>
{{end}}
	</table>
</div>
`
	return template
}

func getMainPage() string {
	template := `
<script>
	function redirect() {
		var sel = document.getElementById('table')
		var value = sel.options[sel.selectedIndex].value;
		if(value == "") {
			return;
		}
		window.location.href='/tables/' + value + '/stats/slowops'
	} 
</script>

<div>
	<h2><i class='fa fa-leaf'></i> Hatchet - MongoDB JSON Log Analyzer</h2>
	<select id='table' class='btn' onchange='redirect()'>
		<option value=''>select a hatchet<option>
{{range $n, $value := .Tables}}
		<option value='{{$value}}'>{{$value}}</option>
{{end}}
	</select>
</div>

<hr/>
<h3>URL</h3>
<ul class="api">
	<li>/</li>
	<li>/tables/{table}</li>
	<li>/tables/{table}/charts/{chart}</li>
	<li>/tables/{table}/logs[?component={str}&context={str}&duration={date},{date}&severity={str}]</li>
	<li>/tables/{table}/logs/slowops[?topN={int}]</li>
	<li>/tables/{table}/stats/slowops[?COLLSCAN={bool}&orderBy={str}]</li>
</ul>

<h3>API</h3>
<ul class="api">
	<li>/api/hatchet/v1.0/tables/{table}/logs[?component={str}&context={str}&duration={date},{date}&severity={str}]</li>
	<li>/api/hatchet/v1.0/tables/{table}/logs/slowops[?topN={int}]</li>
	<li>/api/hatchet/v1.0/tables/{table}/stats/slowops[?COLLSCAN={bool}&orderBy={str}]</li>
</ul>
`
	return template
}
