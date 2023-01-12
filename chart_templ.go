// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"fmt"
	"html/template"
	"time"
)

type NameValue struct {
	Name  string
	Value int
}

func getFooter() string {
	summary := "{{.Summary}}"
	return fmt.Sprintf(`<div class="footer"><img width='32' valign="middle" src='data:image/png;base64,%v'> %v</img></div>`,
		hatchetImage, summary)
}

// GetChartTemplate returns HTML
func GetChartTemplate(attr string, chartType string) (*template.Template, error) {
	html := getContentHTML(attr, chartType)
	if attr == "connections" {
		html += getConnectionsChart()
	} else if attr == "pieChart" {
		html += getPieChart()
	} else {
		html += getOpStatsChart()
	}
	html += "<p/><div id='hatchetChart' align='center' width='100%'/>"
	html += "</body></html>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"descr": func(v OpCount) string {
			dfmt := "2016-01-02T23:59:59"
			d := v.Date + dfmt[len(v.Date):]
			return fmt.Sprintf("%v %v %v %v", v.Op, d, v.Namespace, v.Filter)
		},
		"substr": func(str string, n int) string {
			return str[:n]
		},
		"epoch": func(d string, s string) int64 {
			dfmt := "2016-01-02T23:59:59"
			sdt, _ := time.Parse("2006-01-02T15:04:05", s+dfmt[len(s):])
			dt, _ := time.Parse("2006-01-02T15:04:05", d+dfmt[len(d):])
			return dt.Unix() - sdt.Unix()
		}}).Parse(html)
}

func getOpStatsChart() string {
	return `
<script>
	setChartType();
	google.charts.load('current', {'packages':['corechart']});
	google.charts.setOnLoadCallback(drawChart);

	function drawChart() {
		var data = google.visualization.arrayToDataTable([
			['op', 'secs from origin', 'count', 'op detail'],
{{$sdate := ""}}
{{range $i, $v := .OpCounts}}
{{if eq $i 0}}
	{{$sdate = $v.Date}}
{{end}}
			['', {{epoch $v.Date $sdate}}, {{$v.Count}}, {{descr $v}}],
{{end}}
		]);
		// Set chart options
		var options = {
			'title': 'Ops Stats',
			'hAxis': { textPosition: 'none' },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 600,
			'titleTextStyle': {'fontSize': 20},
			'sizeAxis': {minValue: 0, minSize: 3, maxSize: 3},
			'chartArea': {'width': '90%', 'height': '80%'},
			'legend': { 'position': 'none' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.BubbleChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}

	function refreshOpsStatsChart() {
		sd = document.getElementById('start').value;
		ed = document.getElementById('end').value;
		window.location.href = '/tables/{{.Table}}/charts/slowops?duration=' + sd + 'Z,' + ed + 'Z';
	}
</script>
<input type='datetime-local' id='start' value='{{.Start}}'></input>
<input type='datetime-local' id='end' value='{{.End}}'></input>
<button onClick="refreshOpsStatsChart(); return false;" class="button">Refresh</button>`
}

func getPieChart() string {
	return `
<script>
	setChartType();
	google.charts.load('current', {'packages':['corechart']});
	google.charts.setOnLoadCallback(drawChart);

	function drawChart() {
		var data = google.visualization.arrayToDataTable([
			['Name', 'Value'],
{{range $i, $v := .NameValues}}
			['{{$v.Name}}', {{$v.Value}}],
{{end}}
		]);
		// Set chart options
		var options = {
			'title': '{{.Title}}',
			'height': 480,
			'titleTextStyle': {'fontSize': 20},
			'legend': { 'position': 'left' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.PieChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}
</script>`
}

func getConnectionsChart() string {
	return `
<script>
	setChartType();
	google.charts.load('current', {'packages':['corechart']});
	google.charts.setOnLoadCallback(drawChart);

	function drawChart() {
		var data = google.visualization.arrayToDataTable([
			['IP', 'Accepted', 'Ended'],
{{range $i, $v := .Remote}}
			['{{$v.IP}}', {{$v.Accepted}}, {{$v.Ended}}],
{{end}}
		]);
		// Set chart options
		var options = {
			'title': 'Accepted vs Ended Connections',
			'hAxis': { slantedText:true, slantedTextAngle:15 },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 480,
			'titleTextStyle': {'fontSize': 20},
			'legend': { 'position': 'right' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.ColumnChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}

	function refreshConnsTimeChart() {
		sd = document.getElementById('start').value;
		ed = document.getElementById('end').value;
		window.location.href = '/tables/{{.Table}}/charts/connections?type=time&duration=' + sd + 'Z,' + ed + 'Z';
	}
</script>
{{ if eq .Chart "time" }}
	<input type='datetime-local' id='start' value='{{.Start}}'></input>
	<input type='datetime-local' id='end' value='{{.End}}'></input>
	<button onClick="refreshConnsTimeChart(); return false;" class="button">Refresh</button>
{{ end }}`
}
