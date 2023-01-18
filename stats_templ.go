// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import (
	"fmt"
	"html/template"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetStatsTableTemplate returns HTML
func GetStatsTableTemplate(collscan bool, orderBy string) (*template.Template, error) {
	html := getContentHTML() + getStatsTable(collscan, orderBy) + "</body></html>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"hasPrefix": func(str string, pre string) bool {
			return strings.HasPrefix(str, pre)
		},
		"numPrinter": func(n interface{}) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", ToInt(n))
		}}).Parse(html)
}

func getStatsTable(collscan bool, orderBy string) string {
	checked := ""
	if collscan {
		checked = "checked"
	}
	html := fmt.Sprintf(`
<script>
	function getSlowopsStats() {
		var b = document.getElementById('collscan').checked;
		window.location.href = '/tables/{{.Table}}/stats/slowops?orderBy=%v&COLLSCAN=' + b;
	}
</script>
<p/>`, orderBy)
	html += `<div align='left'><table width='100%'><tr><th>#</th>`
	html += fmt.Sprintf(`<th>op <a class='sort' href='/tables/{{.Table}}/stats/slowops?orderBy=op&COLLSCAN=%v'><i class='fa fa-sort-asc'/></th>`, collscan)
	html += fmt.Sprintf(`<th>namespace <a class='sort' href='/tables/{{.Table}}/stats/slowops?orderBy=ns&order=ASC&COLLSCAN=%v'><i class='fa fa-sort-asc'/></th>`, collscan)
	html += fmt.Sprintf(`<th>count <a class='sort' href='/tables/{{.Table}}/stats/slowops?orderBy=count&COLLSCAN=%v'><i class='fa fa-sort-desc'/></th>`, collscan)
	html += fmt.Sprintf(`<th>avg ms <a class='sort' href='/tables/{{.Table}}/stats/slowops?orderBy=avg_ms&COLLSCAN=%v'><i class='fa fa-sort-desc'/></th>`, collscan)
	html += fmt.Sprintf(`<th>max ms <a class='sort' href='/tables/{{.Table}}/stats/slowops?orderBy=max_ms&COLLSCAN=%v'><i class='fa fa-sort-desc'/></th>`, collscan)
	html += fmt.Sprintf(`<th>total ms <a class='sort' href='/tables/{{.Table}}/stats/slowops?orderBy=total_ms&COLLSCAN=%v'><i class='fa fa-sort-desc'/></th>`, collscan)
	html += fmt.Sprintf(`<th>reslen <a class='sort' href='/tables/{{.Table}}/stats/slowops?orderBy=reslen&COLLSCAN=%v'><i class='fa fa-sort-desc'/></th>`, collscan)
	html += fmt.Sprintf(`<th valign='middle'>index <input type='checkbox' id='collscan' onchange='getSlowopsStats(); return false;' %v></th>`, checked)
	html += `<th>query pattern</th>
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
		{{ if or (eq $value.Index "COLLSCAN") }}
			<td><span style='color:red;'>{{ $value.Index }}</span></td>
		{{ else if (hasPrefix $value.Index "ErrMsg:") }}
			<td align='center'>
				<div class='tooltip'><button class="exclamation"><i class="fa fa-exclamation"></i></button>
					<span class="tooltiptext">{{$value.Index}}</span></div>
				</td>
		{{else}}
			<td>{{ $value.Index }}</td>
		{{end}}
			<td>{{ $value.QueryPattern }}</td>
		</tr>
{{end}}
	</table>
</div>
<p/>`
	return html
}
