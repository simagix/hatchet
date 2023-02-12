// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"fmt"
	"html/template"
	"sort"
)

// GetTablesTemplate returns HTML
func GetTablesTemplate() (*template.Template, error) {
	html := headers + getMainPage() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"getHatchetImage": func() string {
			return HATCHET_PNG
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
      border: 1px solid #ccc;
      vertical-align: middle;
    }
    th {
      background-color: #333;
      color: #fff;
      vertical-align: middle;
      font-size: .8em;
    }
    td {
      vertical-align: middle;
      font-size: .8em;
    }
    .break {
      vertical-align: middle;
      font-size: .8em;
      word-break: break-all;
    }
    tr:nth-child(even) {background-color: #fff;}
    tr:nth-child(odd) {background-color: #f2f2f2;}
    .api {
      font-family: Consolas, monaco, monospace;
    }
    .btn {
      background-color: transparent;
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
    .button {
      font-family: "Trebuchet MS";
      background-color: #4285F4;
      border: none; 
      outline: none;
      color: #f2f2f2;
      padding: 2px 15px;
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
      font-size: 1.6em;
      font-weight: bold;
    }
    h2 {
      font-family: "Trebuchet MS";
      font-size: 1.4em;
      font-weight: bold;
    }
    h3 {
      font-family: "Trebuchet MS";
      font-size: 1.2em;
      font-weight: bold;
    }
    h4 {
      font-family: "Trebuchet MS";
      font-size: 1em;
      font-weight: bold;
    }
    .footer {
      background-color: #fff;
      opacity: .75;
      position: fixed;
      left: 0;
      bottom: 0;
      width: 100%;
      color: #000;
      text-align: left;
      padding: 2px 10px;
    }
    input, select, textarea {
      font-family: "Trebuchet MS";
      appearance: auto;
      background-color: #fff;
      color: #4285F4;
      cursor: pointer;
      border-radius: .25em;
      font-size: 1em;
      padding: 2px 2px;
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
      color: #4285F4;
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
	<button id="title" onClick="javascript:location.href='/'; return false;"
		class="btn" style="float: center;"><i class="fa fa-home"></i> Hatchet</button>

  <button id="logs" onClick="javascript:location.href='/hatchets/{{.Hatchet}}/stats/audit'; return false;"
		class="btn" style="float: left;"><i class="fa fa-shield"></i> Audit</button>
	<button id="stats" onClick="javascript:location.href='/hatchets/{{.Hatchet}}/stats/slowops'; return false;"
		class="btn" style="float: left;"><i class="fa fa-info"></i> Stats</button>
	<button id="logs" onClick="javascript:location.href='/hatchets/{{.Hatchet}}/logs/slowops'; return false;"
		class="btn" style="float: left;"><i class="fa fa-list"></i> Top N</button>
	<button id="search" class="btn" style="float: left;"
		onClick="javascript:location.href='/hatchets/{{.Hatchet}}/logs/all?component=NONE'; return false;">
		<i class="fa fa-search"></i> Search</button>

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
		html += fmt.Sprintf("<option value='/hatchets/{{.Hatchet}}/charts%v'>%v</option>", item.URL, item.Title)
	}

	html += `</select>
	<button id="chart" onClick="javascript:location.href='/hatchets/{{.Hatchet}}/charts/ops?type=stats'; return false;" 
    	class="btn" style="float: right;"><i class="fa fa-bar-chart"></i></button>
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
		window.location.href = '/hatchets/{{.Hatchet}}/charts{{.Chart.URL}}&duration=' + sd + ',' + ed;
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
		window.location.href='/hatchets/' + value + '/stats/audit'
	} 
</script>

<div align='center'>
	<h2><img class='rotate45' width='60' valign="middle" src='data:image/png;base64,{{ getHatchetImage }}'>Hatchet - MongoDB JSON Log Analyzer</img></h2>
	<select id='table' class='hatchet-sel' onchange='redirect()'>
		<option value=''>select a hatchet</option>
{{range $n, $value := .Hatchets}}
		<option value='{{$value}}'>{{$value}}</option>
{{end}}
	</select>
</div>
<hr/>
<h4 align='center'>{{.Version}}</h4>

<h3>Reports</h3>
    <table width='100%'>
      <tr><th></th><th>Title</th><th>Description</th></tr>
      <tr><td align=center><i class="fa fa-shield"></i></td><td>Audit</td><td>Security and audits</td></tr>
      <tr><td align=center><i class="fa fa-bar-chart"></i></td><td>Charts</td><td>Stats charts</td></tr>
      <tr><td align=center><i class="fa fa-search"></i></td><td>Search</td><td>Search logs</td></tr>
      <tr><td align=center><i class="fa fa-info"></i></td><td>Stats</td><td>Slow operations summary</td></tr>
      <tr><td align=center><i class="fa fa-list"></i></td><td>TopN</td><td>Slowest 25 operations</td></tr>
    </table>
<h3>Charts</h3>
    <table width='100%'>
      <tr><th></th><th>Title</th><th>Description</th></tr>`
	size := len(charts) - 1
	tables := make([]Chart, size)
	for k, chart := range charts {
		if k == "instruction" {
			continue
		}
		tables[chart.Index-1] = chart
	}
	for _, chart := range tables {
		template += fmt.Sprintf("<tr><td align=right>%d</td><td>%v</td><td>%v</td></tr>\n",
			chart.Index, chart.Title, chart.Descr)
	}
	template += "</table>"
	template += `<h3>URL</h3>
<ul class="api">
	<li>/</li>
	<li>/hatchets/{hatchet}/charts/{chart}[?type={str}]</li>
	<li>/hatchets/{hatchet}/logs/all[?component={str}&context={str}&duration={date},{date}&severity={str}&limit=[{offset},]{int}]</li>
	<li>/hatchets/{hatchet}/logs/slowops[?topN={int}]</li>
	<li>/hatchets/{hatchet}/stats/slowops[?COLLSCAN={bool}&orderBy={str}]</li>
</ul>

<h3>API</h3>
<ul class="api">
	<li>/api/hatchet/v1.0/hatchets/{hatchet}/logs/all[?component={str}&context={str}&duration={date},{date}&severity={str}&limit=[{offset},]{int}]</li>
	<li>/api/hatchet/v1.0/hatchets/{hatchet}/logs/slowops[?topN={int}]</li>
	<li>/api/hatchet/v1.0/hatchets/{hatchet}/stats/audit</li>
	<li>/api/hatchet/v1.0/hatchets/{hatchet}/stats/slowops[?COLLSCAN={bool}&orderBy={str}]</li>
</ul>
	<div align='center'><hr/><p/>@simagix</div>
`
	template += fmt.Sprintf(`<div class="footer"><img valign="middle" src='data:image/png;base64,%v'/> Ken Chen</div>`, CHEN_ICO)
	return template
}
