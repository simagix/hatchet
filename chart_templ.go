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
	var html string
	if attr == "connections" {
		html = getContentHTML(attr, chartType) + getConnectionsChart() + "</body></html>"
	} else if attr == "pieChart" {
		html = getContentHTML(attr, chartType) + getPieChart() + "</body></html>"
	} else {
		html = getContentHTML(attr, chartType) + getOpCountsChart() + "</body></html>"
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
<script>
	setChartType();
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
			'title': 'Ops Stats',
			'hAxis': { textPosition: 'none' },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 480,
			'titleTextStyle': {'fontSize': 20},
			'chartArea': {'width': '90%', 'height': '80%'},
			'legend': { 'position': 'none' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.BubbleChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}
</script>
<p/>
<div id="hatchetChart" align='center' width='100%'/>
`
	return template
}

func getPieChart() string {
	template := `
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
</script>
<p/>
<div id="hatchetChart" align='center' width='100%'/>
`
	return template
}

func getConnectionsChart() string {
	template := `
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
</script>
<p/>
<div id="hatchetChart" align='center' width='100%'/>
`
	return template
}
