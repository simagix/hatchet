// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"fmt"
	"html/template"
	"time"
)

func getFooter() string {
	summary := "{{.Summary}}"
	return fmt.Sprintf(`<div class="footer"><img valign="middle" src='data:image/png;base64,%v'> %v</img></div>`,
		CHEN_ICO, summary)
}

// GetChartTemplate returns HTML
func GetChartTemplate(chartType string) (*template.Template, error) {
	html := getContentHTML()
	if chartType == BUBBLE_CHART {
		html += getOpStatsChart()
	} else if chartType == PIE_CHART {
		html += getPieChart()
	} else if chartType == BAR_CHART {
		html += getConnectionsChart()
	}
	html += `
	<div style="float: left; width: 100%; clear: left;">
		<input type='datetime-local' id='start' value='{{.Start}}'></input>
		<input type='datetime-local' id='end' value='{{.End}}'></input>
		<button onClick="refreshChart(); return false;" class="button">Refresh</button>
  	</div>
  	<div id='hatchetChart' style="width: 100%; clear: left;"></div>
  
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
{{ if .OpCounts }}
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
	{{$ctype := .Type}}
	{{range $i, $v := .OpCounts}}
		{{if eq $ctype "ops"}}
			[{{$v.Op}}, new Date("{{$v.Date}}"), {{toSeconds $v.Milli}}, '{{descr $v}}', {{$v.Count}}],
		{{else}}
			[{{$v.Op}}, new Date("{{$v.Date}}"), {{$v.Count}}, '{{descr $v}}'],
		{{end}}
	{{end}}
		]);
		// Set chart options
		var options = {
			'backgroundColor': { 'fill': 'transparent' },
			'title': '{{.Chart.Title}}',
			// 'hAxis': { textPosition: 'none' },
			'hAxis': { slantedText: true, slantedTextAngle: 30 },
			'vAxis': {title: '{{.VAxisLabel}}', minValue: 0},
			'width': '100%',
			'height': 480,
			'titleTextStyle': {'fontSize': 20},
			'explorer': { actions: ['dragToZoom', 'rightClickToReset'] },
	{{if eq $ctype "ops"}}
			'sizeAxis': {minValue: 0, minSize: 5, maxSize: 30},
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
</script>
{{else}}
<div align='center' class='btn'><span style='color: red'>no data found</span></div>
{{end}}`
}

func getPieChart() string {
	return `
{{ if .NameValues }}
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
			'backgroundColor': { 'fill': 'transparent' },
			'title': '{{.Chart.Title}}',
			'width': '100%',
			'height': 480,
			'titleTextStyle': {'fontSize': 20},
			'slices': {},
			'legend': { 'position': 'right' } };
		options.slices[data.getSortedRows([{column: 1, desc: true}])[0]] = {offset: 0.1};
		// Instantiate and draw our chart, passing in some options.
		var chart = new google.visualization.PieChart(document.getElementById('hatchetChart'));
		chart.draw(data, options);
	}
</script>
{{else}}
<div align='center' class='btn'><span style='color: red'>no data found</span></div>
{{end}}`
}

func getConnectionsChart() string {
	return `
{{ if .Remote }}
<script>
	setChartType();
	google.charts.load('current', {'packages':['corechart']});
	google.charts.setOnLoadCallback(drawChart);

	function drawChart() {
	{{$ctype := .Type}}
		var data = google.visualization.arrayToDataTable([
	{{if eq $ctype "connections-time"}}
			['Date/Time', 'Connections'],
	{{else}}
			['IP', 'Accepted', 'Ended'],
	{{end}}

	{{range $i, $v := .Remote}}
		{{if eq $ctype "connections-time"}}
			[new Date("{{$v.IP}}"), {{$v.Accepted}}],
		{{else}}
			['{{$v.IP}}', {{$v.Accepted}}, {{$v.Ended}}],
		{{end}}
	{{end}}
		]);
		// Set chart options
		var options = {
			'backgroundColor': { 'fill': 'transparent' },
			'title': '{{.Chart.Title}}',
			'hAxis': { slantedText: true, slantedTextAngle: 30 },
			'vAxis': {title: 'Count', minValue: 0},
			'width': '100%',
			'height': 480,
			'titleTextStyle': {'fontSize': 20},
			'explorer': { actions: ['dragToZoom', 'rightClickToReset'] },
			'legend': { 'position': 'right' } };
		// Instantiate and draw our chart, passing in some options.
	{{if eq $ctype "connections-time"}}
		var chart = new google.visualization.LineChart(document.getElementById('hatchetChart'));
	{{else}}
		var chart = new google.visualization.ColumnChart(document.getElementById('hatchetChart'));
	{{end}}
		chart.draw(data, options);
	}
</script>
{{else}}
<div align='center' class='btn'><span style='color: red'>no data found</span></div>
{{end}}`
}
