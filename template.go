// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"html/template"
	"regexp"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetStatsTemplate returns HTML
func GetStatsTemplate() (*template.Template, error) {
	html := headers + getStatsTable() + "</body>"
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
	html := headers + getLogsTable() + "</body>"
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
	html := headers + getLegacyLogTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		}}).Parse(html)
}

// GetLegacyLogTemplate returns HTML
func GetOpCountsTemplate() (*template.Template, error) {
	html := headers + getOpCountsTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"epoch": func(d string, s string) int64 {
			sdt, _ := time.Parse("2006-01-02T15:04:05", s+":00")
			dt, _ := time.Parse("2006-01-02T15:04:05", d+":00")
			return dt.Unix() - sdt.Unix()
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
    table
    {
    	font-family: Consolas, monaco, monospace;
    	border-collapse:collapse;
    	min-width:600px;
    }
    caption
    {
    	caption-side:top;
    	font-weight:bold;
    	font-style:italic;
    	margin:2px;
    }
    table, th, td
    {
		border: 1px solid gray;
		vertical-align: top;
    }
    th, td
    {
    	padding:2px;
    	vertical-align: top;
    }
    th
    {
      background-color: #ddd;
      font-weight:bold;
    }
    tr:nth-child(even) {background-color: #f2f2f2;}
    tr:nth-child(odd) {background-color: #fff;}
	.btn {
	  background-color: #fff;
	  border: none;
	  outline:none;
	  color: #4285F4;
	  padding: 5px 30px;
	  cursor: pointer;
	  font-size: 20px;
	}
	.btn:hover {
	  color: #DB4437;
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
<div align='center'>
	<button id="stats" onClick="javascript:location.href='/tables/{{.Table}}/stats/slowops'; return false;"
		class="btn" style="float: left;"><i class="fa fa-info"></i> Stats</button>
	<button id="stats" onClick="javascript:location.href='/tables/{{.Table}}/charts/slowops'; return false;"
		class="btn" style="float: center;"><i class="fa fa-pie-chart"></i> Chart</button>
	<button id="stats" onClick="javascript:location.href='/tables/{{.Table}}/logs/slowops'; return false;"
		class="btn" style="float: right;"><i class="fa fa-database"></i> Top N</button>
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

func getOpCountsTable() string {
	template := `
<script>
	google.charts.load('current', {'packages':['corechart']});
	google.charts.setOnLoadCallback(drawChart);
	function drawChart() {
		var data = google.visualization.arrayToDataTable([
			['op', 'secs+', 'count', 'op detail', 'count'],
{{$sdate := ""}}
{{range $i, $v := .OpCounts}}
{{if eq $i 0}}
	{{$sdate = $v.Date}}
{{end}}
			['{{$v.Op}}', {{epoch $v.Date $sdate}}, {{$v.Count}}, '{{$v.Date}} {{$v.Namespace}} {{$v.Filter}}', {{$v.Count}}],
{{end}}
		]);
		// Set chart options
		var options = {
			'title':'Ops Counts Chart ({{.Table}})',
			'hAxis': { textPosition: 'none' },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 600,
			'titleTextStyle': {'fontSize': 20},
			'chartArea': {'width': '90%', 'height': '90%'},
			'legend': { 'position': 'none' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.BubbleChart(document.getElementById('OpsCounts'));
		chart.draw(data, options);
	}
</script>

<div id="OpsCounts" align='center' width='100%'>
</div>
`
	return template
}
