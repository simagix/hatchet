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
func GetChartTemplate(chartType string) (*template.Template, error) {
	html := getContentHTML()
	if chartType == "ops" || chartType == "slowops" {
		html += getOpStatsChart()
	} else if chartType == "slowops-counts" || chartType == "connections-accepted" {
		html += getPieChart()
	} else if chartType == "connections-time" || chartType == "connections-total" {
		html += getConnectionsChart()
	}
	html += `
		<input type='datetime-local' id='start' value='{{.Start}}'></input>
		<input type='datetime-local' id='end' value='{{.End}}'></input>
		<button onClick="refreshChart(); return false;" class="button">Refresh</button>
		<div id='hatchetChart' align='center' width='100%'/>
		</body></html>`
	return template.New("hatchet").Funcs(template.FuncMap{
		"descr": func(v OpCount) template.HTML {
			if v.Filter == "" {
				return template.HTML(v.Namespace)
			}
			str := fmt.Sprintf("%v, QP: %v", v.Namespace, v.Filter)
			return template.HTML(str)
		},
		"toSeconds": func(n float64) float64 {
			return n / 1000
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
{{if eq .Type "ops"}}
			['op', 'date/time', 'duration (seconds)', 'ns', 'counts'],
{{else}}
			['op', 'date/time', 'count', 'ns/filter'],
{{end}}

{{$sdate := ""}}
{{$ctype := .Type}}
{{range $i, $v := .OpCounts}}
{{if eq $i 0}}
	{{$sdate = $v.Date}}
{{end}}

{{if eq $ctype "ops"}}
			[{{$v.Op}}, new Date("{{$v.Date}}"), {{toSeconds $v.Milli}}, '{{descr $v}}', {{$v.Count}}],
{{else}}
			[{{$v.Op}}, new Date("{{$v.Date}}"), {{$v.Count}}, '{{descr $v}}'],
{{end}}
{{end}}
		]);
		// Set chart options
		var options = {
			'title': '{{.Chart.Title}}',
			'hAxis': { textPosition: 'none' },
			'vAxis': {title: '{{.VAxisLabel}}', minValue: 0},
			'height': 600,
			'titleTextStyle': {'fontSize': 20},
{{if eq $ctype "ops"}}
			'sizeAxis': {minValue: 0, minSize: 5, maxSize: 15},
{{else}}
			'sizeAxis': {minValue: 0, minSize: 5, maxSize: 5},
{{end}}
			'chartArea': {'width': '85%', 'height': '80%'},
			'tooltip': { 'isHtml': false },
			'legend': { 'position': 'none' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.BubbleChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}
</script>`
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
			'title': '{{.Chart.Title}}',
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
			'title': '{{.Chart.Title}}',
			'hAxis': { slantedText:true, slantedTextAngle:15 },
			'vAxis': {title: 'Count', minValue: 0},
			'height': 480,
			'titleTextStyle': {'fontSize': 20},
			'legend': { 'position': 'bottom' } };
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.ColumnChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}
</script>`
}
