/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * template.go
 */

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
    :root {
      --text-color: #2C5234;
      --header-color: #C1D8C5;
      --row-color: #E8F1E9;
      --background-color: #F3F7F4;
      --accent-color-1: #7BAF9B;
      --accent-color-2: #9FCCB3;
      --accent-color-3: #5E8961;
      --border-color: #DDD;
    }
  	body {
      font-family: Helvetica, Arial, sans-serif;
      margin-top: 10px;
      margin-bottom: 10px;
      margin-right: 10px;
      margin-left: 10px;
	    background-color: var(--background-color);
    }
    table {
      border-collapse:collapse;
      min-width: 300px;
    }
    caption {
      caption-side: top;
      font-size: 1em;
	    text-align: left;
    }
    table, th, td {
      border: 1px solid var(--border-color);
      vertical-align: middle;
    }
    th {
      background-color: var(--header-color);
      color: var(--text-color);
      font-weight: bold;
      padding: 0.1rem;
      font-size: 0.9em;
      text-align: left;    
    }
    td {
      background-color: var(--row-color);
      padding: 0.1rem;
      font-size: 0.9em;
    }
    tr:nth-child(even) td {
      background-color: white;
    }
    .break {
      vertical-align: middle;
      font-size: .8em;
      word-break: break-all;
    }
    table a:link {
      color: var(--text-color);
      text-decoration: none;
    }
    table a:visited {
      color: var(-accent-color-1);
      text-decoration: none;
    }
    table a:hover {
      color: red;
      text-decoration: none;
    }
    ul, ol {
      #font-family: Consolas, monaco, monospace;
      font-size: .8em;
    }
    .btn {
      background-color: transparent;
      border: none;
      outline:none;
      color: var(--accent-color-3);
      padding: 2px 2px;
      cursor: pointer;
      font-size: 16px;
      font-weight: bold;
      border-radius: .25em;
    }
    .btn:hover {
      background-color: var(--border-color);
      color: #DB4437;
    }
    .button { 
      background-color: var(--text-color);
      border: none; 
      outline: none;
      color: var(--background-color);
      padding: 3px 15px;
      margin: 0px 10px;
      cursor: pointer;
      font-size: 14px;
      font-weight: bold;
      border-radius: 3px;
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
      font-size: 1.6em;
      font-weight: bold;
    }
    h2 {
      font-size: 1.4em;
      font-weight: bold;
    }
    h3 {
      font-size: 1.2em;
      font-weight: bold;
    }
    h4 {
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
      color: #505950;
      border-radius: .25em;
      font-size: .9em;
      #padding: 5px 5px;
    }
    .rotate23:hover {
      -webkit-transform: rotate(23deg);
      -moz-transform: rotate(23deg);
      -o-transform: rotate(23deg);
      -ms-transform: rotate(23deg);
      transform: rotate(23deg);
    }
    input[type="checkbox"] {
      accent-color: red;
    }
    .sort {
      color: #FFF;
    }
    .sort:hover {
      color: #DB4437;
    }
    .summary {
      font-family: Consolas, monaco, monospace;
	    background-color: #111;
      color: var(--border-color);
	    padding: .5rem;
	    margin: .5rem;
      font-size: .8em;
    }
    #loading {
      position: fixed;
      top: 0;
      left: 0;
      bottom: 0;
      right: 0;
      background-color: rgba(0, 0, 0, 0.5);
      z-index: 9999;
	  display: none;
    }
    .spinner {
      border: 5px solid #f3f3f3;
      border-top: 5px solid #3498db;
      border-radius: 50%;
      width: 50px;
      height: 50px;
      animation: spin 2s linear infinite;
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
    }
    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }
  </style>
  <script>
    function loadData(url) {
    	var loading = document.getElementById('loading');
    	loading.style.display = 'block';
    	fetch(url)
        	.then(response => response.text())
        	.then(data => {
      			loading.style.display = 'none';
				document.open();
				document.write(data);
				document.close();
        	})
        	.catch(error => {
      			loading.style.display = 'none';
        	});
    }
  </script>
</head>
<body>
  <div id="loading">
    <div class="spinner"></div>
  </div>
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
		loadData(value);
	}
</script>
<div align='center'>
	<div style="float: left; margin-right: 10px;">
	  <button id="title" onClick="javascript:location.href='/'; return false;"
		class="btn"><i class="fa fa-home"></i></button>Hatchet</div>

  <div style="float: left; margin-right: 10px;">
  	<button id="logs" onClick="javascript:loadData('/hatchets/{{.Hatchet}}/stats/audit'); return false;"
		class="btn"><i class="fa fa-shield"></i></button>Audit</div>
  <div style="float: left; margin-right: 10px;">
  	<button id="stats" onClick="javascript:loadData('/hatchets/{{.Hatchet}}/stats/slowops'); return false;"
		class="btn"><i class="fa fa-info"></i></button>Stats</div>
  <div style="float: left; margin-right: 10px;">
  	<button id="logs" onClick="javascript:loadData('/hatchets/{{.Hatchet}}/logs/slowops'); return false;"
		class="btn"><i class="fa fa-list"></i></button>Top N</div>
  <div style="float: left; margin-right: 10px;">
  	<button id="search" onClick="javascript:loadData('/hatchets/{{.Hatchet}}/logs/all?component=NONE'); return false;"
    	class="btn"><i class="fa fa-search"></i></button>Search</div>

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
	<button id="chart" onClick="javascript:loadData('/hatchets/{{.Hatchet}}/charts/ops?type=stats'); return false;" 
    	class="btn" style="float: right;"><i class="fa fa-bar-chart"></i></button>
</div>
<script>
	function setChartType() {
		var sel = document.getElementById('nextChart')
		sel.selectedIndex = {{.Chart.Index}};
	}

	function refreshChart() {
		var sd = document.getElementById('start').value;
		var ed = document.getElementById('end').value;
		loadData('/hatchets/{{.Hatchet}}/charts{{.Chart.URL}}&duration=' + sd + ',' + ed);
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
		loadData('/hatchets/' + value + '/stats/audit'); 
	} 
</script>

<div align='center'>
	<h2><img class='rotate23' width='60' valign="middle" src='data:image/png;base64,{{ getHatchetImage }}'>Hatchet - MongoDB JSON Log Analyzer</img></h2>
	<select id='table' class='hatchet-sel' onchange='javascript:redirect(); return false'>
		<option value=''>select a hatchet</option>
{{range $n, $value := .Hatchets}}
		<option value='{{$value}}'>{{$value}}</option>
{{end}}
	</select>
</div>
<hr/>
<h3>Reports</h3>
    <table width='100%'>
      <tr><th></th><th>Title</th><th>Description</th></tr>
      <tr><td align=center><i class="fa fa-shield"></i></td><td>Audit</td><td>Display information on security audits and performance metrics</td></tr>
      <tr><td align=center><i class="fa fa-bar-chart"></i></td><td>Charts</td><td>A number of charts are available for security audits and performance metrics</td></tr>
      <tr><td align=center><i class="fa fa-search"></i></td><td>Search</td><td>Powerful log searching function with key metrics highlighted</td></tr>
      <tr><td align=center><i class="fa fa-info"></i></td><td>Stats</td><td>Summary of slow operational query patterns and duration</td></tr>
      <tr><td align=center><i class="fa fa-list"></i></td><td>TopN</td><td>Display the slowest 23 operation logs</td></tr>
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
	<li>/api/hatchet/v1.0/mongodb/{version}/drivers/{driver}?compatibleWith={driver version}</li>
</ul>
<h4 align='center'><hr/>{{.Version}}</h4>
`
	template += fmt.Sprintf(`<div class="footer"><img valign="middle" src='data:image/png;base64,%v'/> Ken Chen</div>`, CHEN_ICO)
	return template
}
