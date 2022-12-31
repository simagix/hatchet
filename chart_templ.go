// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"html/template"
	"time"
)

// GetChartTemplate returns HTML
func GetChartTemplate(attr string) (*template.Template, error) {
	var html string
	if attr == "connections" {
		html = headers + menuHTML + getConnectionsChart() + "</body>"
	} else if attr == "accepted_conns" {
		html = headers + menuHTML + getAcceptedConnsChart() + "</body>"
	} else {
		html = headers + menuHTML + getOpCountsChart() + "</body>"
	}
	return template.New("hatchet").Funcs(template.FuncMap{
		"epoch": func(d string, s string) int64 {
			sdt, _ := time.Parse("2006-01-02T15:04:05", s+":00")
			dt, _ := time.Parse("2006-01-02T15:04:05", d+":00")
			return dt.Unix() - sdt.Unix()
		}}).Parse(html)
}

func getOpCountsChart() string {
	template := `
<div id="OpCounts" align='center' width='100%'/>
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
			'title':'Ops Stats',
			'hAxis': { textPosition: 'none' },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 600,
			'titleTextStyle': {'fontSize': 20},
			'chartArea': {'width': '90%', 'height': '80%'},
			'legend': { 'position': 'none' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.BubbleChart(document.getElementById('OpCounts'));
		chart.draw(data, options);
	}
</script>
`
	return template
}

func getAcceptedConnsChart() string {
	template := `
<div id="OpCounts" align='center' width='100%'/>
<script>
	google.charts.load('current', {'packages':['corechart']});
	google.charts.setOnLoadCallback(drawChart);

	function drawChart() {
		var data = google.visualization.arrayToDataTable([
			['IP', 'Conns'],
{{range $i, $v := .Remote}}
			['{{$v.IP}}', {{$v.Conns}}],
{{end}}
		]);
		// Set chart options
		var options = {
			'title':'Accepted Connections',
			'hAxis': { slantedText:true, slantedTextAngle:15 },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 600,
			'titleTextStyle': {'fontSize': 20},
			'legend': { 'position': 'right' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.PieChart(document.getElementById('OpCounts'));
		chart.draw(data, options);
	}
</script>
`
	return template
}

func getConnectionsChart() string {
	template := `
<div id="OpCounts" align='center' width='100%'/>
<script>
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
			'title':'Accepted vs Ended Connections',
			'hAxis': { slantedText:true, slantedTextAngle:15 },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 600,
			'titleTextStyle': {'fontSize': 20},
			'legend': { 'position': 'right' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.ColumnChart(document.getElementById('OpCounts'));
		chart.draw(data, options);
	}
</script>
`
	return template
}
