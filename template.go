// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"fmt"
	"html/template"
	"sort"
)

const hatchetImage = `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`

// GetTablesTemplate returns HTML
func GetTablesTemplate() (*template.Template, error) {
	html := headers + getMainPage() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"getHatchetImage": func() string {
			return hatchetImage
		},
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
	  background-color: #f2f2f2;
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
      vertical-align: middle;
    }
    th {
      background-color: #888;
      color: #fff;
      vertical-align: middle;
      font-size: .9em;
    }
    td {
      padding:2px;
      vertical-align: middle;
      font-size: 11px;
    }
    tr:nth-child(even) {background-color: #f2f2f2;}
    tr:nth-child(odd) {background-color: #fff;}
    .api {
      font-family: Consolas, monaco, monospace;
    }
    .btn {
      background-color: #f2f2f2;
      border: none;
      outline:none;
      color: #4285F4;
      padding: 5px 10px;
      cursor: pointer;
      font-size: 16px;
    }
    .btn:hover {
      color: #DB4437;
	  cursor: hand;
    }
    .btn2 {
      background-color: #f2f2f2;
      color: #4285F4;
      cursor: pointer;
      font-size: .8em;
      padding: 5px;
      border-radius: .25em;
    }
    .btn2:hover {
      color: #DB4437;
	    cursor: hand;
    }
    .button {
      font-family: "Trebuchet MS";
      background-color: #4285F4;
      border: none; 
      outline: none;
      color: #f2f2f2;
      padding: 5px 15px;
      margin: 2px 10px;
      cursor: pointer;
      font-size: 1em;
      font-weight: bold;
      border-radius: .25em;
    }
    .exclamation {
      background: none;
      color: red;
      border: none;
      outline: none;
      padding: 5px 10px;
      margin: 2px 2px;
      font-size: 1em;
      border-radius: .25em;
    }
    .tooltip {
      position: relative;
      display: inline-block;
    }
    .tooltip .tooltiptext {
      visibility: hidden;
      width: 200px;
      background-color: #555;
      color: #fff;
      text-align: left;
      border-radius: 6px;
      padding: 5px 5px;
      position: absolute;
      z-index: 1;
      bottom: 125%;
      left: 50%;
      margin-left: -100px;
      opacity: 0;
      transition: opacity 0.3s;
    }
    .tooltip .tooltiptext::after {
      content: "";
      position: absolute;
      top: 100%;
      left: 50%;
      margin-left: -5px;
      border-width: 5px;
      border-style: solid;
      border-color: #555 transparent transparent transparent;
    }
    .tooltip:hover .tooltiptext {
      visibility: visible;
      opacity: 1;
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
    .footer {
      background-color: #4285F4;
      opacity: .8;
      position: fixed;
      left: 0;
      bottom: 0;
      width: 100%;
      color: #fff;
      text-align: center;
    }
    input, select, textarea {
      font-family: "Trebuchet MS";
      appearance: auto;
      background-color: #fff;
      color: #4285F4;
      cursor: pointer;
      border-radius: .25em;
      font-size: 1em;
      padding: 5px 5px;
    }
    .hatchet-sel {
      font-size: 1.25em;
      padding: 10px 20px;
    }
    .rotate45:hover {
      -webkit-transform: rotate(45deg);
      -moz-transform: rotate(45deg);
      -o-transform: rotate(45deg);
      -ms-transform: rotate(45deg);
      transform: rotate(45deg);
    }
    input[type="checkbox"] {
      accent-color: red;
    }
    .sort {
      color: #fff;
    }
    .sort:hover {
      color: #DB4437;
	  cursor: hand;
    }
  </style>
</head>

<body>
`

func getContentHTML() string {
	html := headers
	html += `
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
	<button id="stats" onClick="javascript:location.href='/tables/{{.Table}}/stats/slowops'; return false;"
		class="btn" style="float: left;"><i class="fa fa-info"></i> Stats</button>
	<button id="logs" onClick="javascript:location.href='/tables/{{.Table}}/logs/slowops'; return false;"
		class="btn" style="float: left;"><i class="fa fa-list"></i> Top N</button>

	<button id="title" onClick="javascript:location.href='/'; return false;"
		class="btn" style="float: center;"><i class="fa fa-home"></i> Hatchet</button>

	<button id="search" onClick="javascript:location.href='/tables/{{.Table}}/logs?component=NONE'; return false;"
		class="btn" style="float: right;"><i class="fa fa-search"></i></button>
	<select id='nextChart' style="float: right;" onchange='gotoChart()'>`
	items := []Chart{}
	for _, chart := range charts {
		items = append(items, chart)
	}
	sort.Slice(items, func(i int, j int) bool {
		return items[i].Index < items[j].Index
	})

	html += "<option value=''>select a chart</option>"
	for i, item := range items {
		if i == 0 {
			continue
		}
		html += fmt.Sprintf("<option value='/tables/{{.Table}}/charts%v'>%v</option>", item.URL, item.Label)
	}

	html += `</select>
	<button class="btn" style="float: right;"><i class="fa fa-bar-chart"></i></button>
</div>
<p/>
<script>
	function setChartType() {
		var sel = document.getElementById('nextChart')
		sel.selectedIndex = {{.Chart.Index}};
	}

	function refreshChart() {
		var sd = document.getElementById('start').value;
		var ed = document.getElementById('end').value;
		window.location.href = '/tables/{{.Table}}/charts{{.Chart.URL}}&duration=' + sd + 'Z,' + ed + 'Z';
	}
</script>
`
	html += getFooter()
	return html
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

<div align='center'>
	<h2><img class='rotate45' width='60' valign="middle" src='data:image/png;base64,{{ getHatchetImage }}'>Hatchet - MongoDB JSON Log Analyzer</img></h2>
	<select id='table' class='hatchet-sel' onchange='redirect()'>
		<option value=''>select a hatchet</option>
{{range $n, $value := .Tables}}
		<option value='{{$value}}'>{{$value}}</option>
{{end}}
	</select>
</div>
<hr/>
<h4 align='center'>{{.Version}}</h4>
<h3>URL</h3>
<ul class="api">
	<li>/</li>
	<li>/tables/{table}</li>
	<li>/tables/{table}/charts/{chart}[?type={str}]</li>
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
<div class="footer">created by Ken Chen (https://github.com/simagix/hatchet)</div>
`
	return template
}
