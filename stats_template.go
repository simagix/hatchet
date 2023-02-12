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
func GetStatsTableTemplate(collscan bool, orderBy string, download string) (*template.Template, error) {
	html := headers
	if download == "" {
		html = getContentHTML() + getFooter()
	}
	html += getStatsTable(collscan, orderBy, download) + "</body></html>"
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

func getStatsTable(collscan bool, orderBy string, download string) string {
	checked := ""
	if collscan {
		checked = "checked"
	}
	html := fmt.Sprintf(`
<script>
	function getSlowopsStats() {
		var b = document.getElementById('collscan').checked;
		window.location.href = '/hatchets/{{.Hatchet}}/stats/slowops?orderBy=%v&COLLSCAN=' + b;
	}
	
	function downloadStats() {
        anchor = document.createElement('a');
        anchor.download = '{{.Hatchet}}_stats.html';
        anchor.href = '/hatchets/{{.Hatchet}}/stats/slowops?type=stats&download=true';
        anchor.dataset.downloadurl = ['text/html', anchor.download, anchor.href].join(':');
        anchor.click();
    }
</script>`, orderBy)
	asc := "<i class='fa fa-sort-asc'/>"
	desc := "<i class='fa fa-sort-desc'/>"
	html += `<div align='left'>`
	if download == "" {
		html += `<button id="download" onClick="downloadStats(); return false;"
			class="btn" style="float: right;"><i class="fa fa-download"></i></button>`
	} else {
		html += "<div align='center'>{{.Summary}}</div>"
		asc = ""
		desc = ""
	}
	html += `<table width='100%'><tr><th>#</th>`
	html += fmt.Sprintf(`<th>op <a class='sort' href='/hatchets/{{.Hatchet}}/stats/slowops?orderBy=op&COLLSCAN=%v'>%v</th>`, collscan, asc)
	html += fmt.Sprintf(`<th>namespace <a class='sort' href='/hatchets/{{.Hatchet}}/stats/slowops?orderBy=ns&order=ASC&COLLSCAN=%v'>%v</th>`, collscan, asc)
	html += fmt.Sprintf(`<th>count <a class='sort' href='/hatchets/{{.Hatchet}}/stats/slowops?orderBy=count&COLLSCAN=%v'>%v</th>`, collscan, desc)
	html += fmt.Sprintf(`<th>avg ms <a class='sort' href='/hatchets/{{.Hatchet}}/stats/slowops?orderBy=avg_ms&COLLSCAN=%v'>%v</th>`, collscan, desc)
	html += fmt.Sprintf(`<th>max ms <a class='sort' href='/hatchets/{{.Hatchet}}/stats/slowops?orderBy=max_ms&COLLSCAN=%v'>%v</th>`, collscan, desc)
	html += fmt.Sprintf(`<th>total ms <a class='sort' href='/hatchets/{{.Hatchet}}/stats/slowops?orderBy=total_ms&COLLSCAN=%v'>%v</th>`, collscan, desc)
	html += fmt.Sprintf(`<th>reslen <a class='sort' href='/hatchets/{{.Hatchet}}/stats/slowops?orderBy=reslen&COLLSCAN=%v'>%v</th>`, collscan, desc)
	if download == "" {
		html += fmt.Sprintf(`<th valign='middle'>index <input type='checkbox' id='collscan' onchange='getSlowopsStats(); return false;' %v></th>`, checked)
	} else {
		html += "<th valign='middle'>index</th>"
	}
	html += `<th>query pattern</th>
		</tr>
{{range $n, $value := .Ops}}
		<tr>
			<td align='right'>{{ add $n 1 }}</td>
			<td class='break'>{{ $value.Op }}</td>
			<td class='break'>{{ $value.Namespace }}</td>
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
			<td class='break'>{{ $value.QueryPattern }}</td>
		</tr>
{{end}}
	</table>
	</div>
	<div align='center'><hr/><p/>@simagix</div>
</div>`
	return html
}

// GetAuditTablesTemplate returns HTML
func GetAuditTablesTemplate() (*template.Template, error) {
	html := headers + getContentHTML() + getFooter()
	html += `
{{if hasData .Data "exception"}}
	<h3><i class="fa fa-shield"></i> Exceptions<button class='btn'
		onClick="javascript:location.href='/hatchets/{{.Hatchet}}/logs/all?severity=W'; return false;">
		<i class='fa fa-search'></i></button></h3>
	<table>
		<tr><th></th><th>Severity</th><th>Counts</th></tr>
	{{range $n, $val := index .Data "exception"}}
		<tr><td align=right>{{add $n 1}}</td><td>{{$val.Name}}</td><td align=right>{{numPrinter $val.Value}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "failed"}}
	<h3><i class="fa fa-shield"></i> Failed Operations</h3>
	<table>
		<tr><th></th><th>Failed Operations</th><th>Counts</th></tr>
	{{$name := .Hatchet}}
	{{range $n, $val := index .Data "failed"}}
		<tr><td align=right>{{add $n 1}}</td>
			<td>
				{{$val.Name}} <button class='btn'onClick="javascript:location.href='/hatchets/{{$name}}/logs/all?context={{$val.Name}}'; return false;">
				<i class='fa fa-search'></i></button>
			</td>
			<td align=right>{{numPrinter $val.Value}}</td>
		</tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "op"}}
	<h3><i class="fa fa-shield"></i> Operations Stats<button class='btn'
		onClick="javascript:location.href='/hatchets/{{.Hatchet}}/stats/slowops'; return false;">
		<i class='fa fa-info'></i></button></h3>
	<table>
		<tr><th></th><th>Operation</th><th>Counts</th></tr>
	{{range $n, $val := index .Data "op"}}
		<tr><td align=right>{{add $n 1}}</td><td>{{$val.Name}}</td><td align=right>{{numPrinter $val.Value}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "ip"}}
	<h3><i class="fa fa-shield"></i> Connected Clients<button class='btn'
		onClick="javascript:location.href='/hatchets/{{.Hatchet}}/charts/connections?type=accepted'; return false;">
		<i class='fa fa-bar-chart'></i></button></h3>
	<table>
		<tr><th></th><th>IP</th><th>Accepted Connections</th></tr>
	{{range $n, $val := index .Data "ip"}}
		<tr><td align=right>{{add $n 1}}</td><td>{{$val.Name}}</td><td align=right>{{numPrinter $val.Value}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "ns"}}
	<h3><i class="fa fa-shield"></i> Namespaces<button class='btn'
		onClick="javascript:location.href='/hatchets/{{.Hatchet}}/charts/ops?type=stats'; return false;">
		<i class='fa fa-bar-chart'></i></button></h3>
	<table>
		<tr><th></th><th>Namespace</th><th>Counts</th></tr>
	{{range $n, $val := index .Data "ns"}}
		<tr><td align=right>{{add $n 1}}</td><td>{{$val.Name}}</td><td align=right>{{numPrinter $val.Value}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "reslen"}}
	<h3><i class="fa fa-shield"></i> Response Length<button class='btn'
		onClick="javascript:location.href='/hatchets/{{.Hatchet}}/charts/reslen?type=ips'; return false;">
		<i class='fa fa-bar-chart'></i></button></h3>
	<table>
		<tr><th></th><th>IP</th><th>Response Length (bytes)</th></tr>
	{{range $n, $val := index .Data "reslen"}}
		<tr><td align=right>{{add $n 1}}</td><td>{{$val.Name}}</td><td align=right>{{numPrinter $val.Value}}</td></tr>
	{{end}}
	</table>
{{end}}
	<div align='center'><hr/><p/>@simagix</div>
	`
	html += "</body></html>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"hasData": func(data map[string][]NameValue, key string) bool {
			return len(data[key]) > 0
		},
		"numPrinter": func(n interface{}) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", ToInt(n))
		}}).Parse(html)
}
