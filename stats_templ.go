// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"html/template"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetStatsTemplate returns HTML
func GetStatsTemplate() (*template.Template, error) {
	html := headers + menuHTML + getStatsTable() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"numPrinter": func(n interface{}) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", ToInt(n))
		}}).Parse(html)
}

func getStatsTable() string {
	template := `
<div><p/>{{.Summary}}</div>
<div align='left'>
	<table width='100%'>
		<tr>
			<th>#</th>
			<th>op <a href='/tables/{{.Table}}/stats/slowops?orderBy=op'><i class='fa fa-sort-asc'/></th>
			<th>namespace <a href='/tables/{{.Table}}/stats/slowops?orderBy=ns&order=ASC'><i class='fa fa-sort-asc'/></th>
			<th>count <a href='/tables/{{.Table}}/stats/slowops?orderBy=count'><i class='fa fa-sort-desc'/></th>
			<th>avg ms <a href='/tables/{{.Table}}/stats/slowops?orderBy=avg_ms'><i class='fa fa-sort-desc'/></th>
			<th>max ms <a href='/tables/{{.Table}}/stats/slowops?orderBy=max_ms'><i class='fa fa-sort-desc'/></th>
			<th>total ms <a href='/tables/{{.Table}}/stats/slowops?orderBy=total_ms'><i class='fa fa-sort-desc'/></th>
			<th>reslen <a href='/tables/{{.Table}}/stats/slowops?orderBy=reslen'><i class='fa fa-sort-desc'/></th>
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
