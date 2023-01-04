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

func getFooter(attr string, chartType string) string {
	str := fmt.Sprintf(`
<script>
	function setChartType() {
		var sel = document.getElementById('nextChart')
		var attr = '%v';
		var chartType = '%v';

		if(attr == "slowops" && (chartType == "" || chartType == "stats")) {
			sel.selectedIndex = 1;
		} else if(attr == "pieChart" && chartType == "counts") {
			sel.selectedIndex = 2;
		} else if(attr == "pieChart" && chartType == "accepted") {
			sel.selectedIndex = 3;
		} else if(attr == "connections" && chartType == "time") {
			sel.selectedIndex = 4;
		} else if(attr == "connections" && chartType == "total") {
			sel.selectedIndex = 5;
		}
	}
</script>`, attr, chartType);
	str += fmt.Sprintf(`<div class="footer"><img width='32' valign="middle" src='data:image/png;base64,%v'>Hatchet</img></div>`,
		hatchetImage)
	return str
}

// GetChartTemplate returns HTML
func GetChartTemplate(attr string, chartType string) (*template.Template, error) {
	var html string
	if attr == "connections" {
		html = headers + menuHTML + getFooter(attr, chartType) + getConnectionsChart() + "</body></html>"
	} else if attr == "pieChart" {
		html = headers + menuHTML + getFooter(attr, chartType) + getPieChart() + "</body></html>"
	} else {
		html = headers + menuHTML + getFooter(attr, chartType) + getOpCountsChart() + "</body></html>"
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
			'height': 400,
			'titleTextStyle': {'fontSize': 20},
			'chartArea': {'width': '90%', 'height': '80%'},
			'legend': { 'position': 'none' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.BubbleChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}
</script>
<div><p/>{{.Summary}}</div>
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
			'hAxis': { slantedText:true, slantedTextAngle:15 },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 400,
			'titleTextStyle': {'fontSize': 20},
			'legend': { 'position': 'right' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.PieChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}
</script>
<div><p/>{{.Summary}}</div>
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
			'height': 400,
			'titleTextStyle': {'fontSize': 20},
			'legend': { 'position': 'right' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.ColumnChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}
</script>
<div><p/>{{.Summary}}</div>
<div id="hatchetChart" align='center' width='100%'/>
`
	return template
}
