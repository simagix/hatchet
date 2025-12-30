/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * stats_template.go
 */

package hatchet

import (
	"fmt"
	"html/template"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const MIN_MONGO_VER = "5.0"

// GetStatsTableTemplate returns HTML
func GetStatsTableTemplate(collscan bool, orderBy string, download string) (*template.Template, error) {
	html := headers
	if download == "" {
		html = getContentHTML()
	}
	html += getStatsTable(collscan, orderBy, download)
	if download == "" {
		html += "</div><!-- end content-container -->"
	}
	html += "</body></html>"
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
		},
		"getMarkerHTML": func(marker int) template.HTML {
			return template.HTML(GetMarkerHTML(marker))
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
		loadData('/hatchets/{{.Hatchet}}/stats/slowops?orderBy=%v&COLLSCAN='+b);
	}
	function downloadStats() {
        anchor = document.createElement('a');
        anchor.download = '{{.Hatchet}}_stats.html';
        anchor.href = '/hatchets/{{.Hatchet}}/stats/slowops?type=stats&download=true';
        anchor.dataset.downloadurl = ['text/html', anchor.download, anchor.href].join(':');
        anchor.click();
    }
	function toggleStatsJson(id) {
		var row = document.getElementById('json-stats-' + id);
		var btn = document.getElementById('btn-stats-' + id);
		if (row.style.display === 'none' || row.style.display === '') {
			row.style.display = 'table-row';
			btn.classList.add('active');
			// Format on first open
			var indexEl = document.getElementById('index-content-' + id);
			var patternEl = document.getElementById('pattern-content-' + id);
			if (indexEl && !indexEl.dataset.formatted) {
				indexEl.innerHTML = formatJsonContent(indexEl.textContent);
				indexEl.dataset.formatted = 'true';
			}
			if (patternEl && !patternEl.dataset.formatted) {
				patternEl.innerHTML = formatJsonContent(patternEl.textContent);
				patternEl.dataset.formatted = 'true';
			}
		} else {
			row.style.display = 'none';
			btn.classList.remove('active');
		}
	}
	function syntaxHighlightStats(json) {
		json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
		return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
			var cls = 'json-number';
			if (/^"/.test(match)) {
				if (/:$/.test(match)) {
					cls = 'json-key';
				} else {
					cls = 'json-string';
				}
			} else if (/true|false/.test(match)) {
				cls = 'json-boolean';
			} else if (/null/.test(match)) {
				cls = 'json-null';
			}
			return '<span class="' + cls + '">' + match + '</span>';
		});
	}
	function formatJsonContent(text) {
		try {
			var json = JSON.parse(text);
			return syntaxHighlightStats(JSON.stringify(json, null, 2));
		} catch(e) {
			// Try to fix unquoted keys (Go-style map format)
			try {
				// Add quotes around unquoted keys: word: or $word:
				var fixed = text.replace(/([{,]\s*)(\$?[a-zA-Z_][a-zA-Z0-9_.]*)\s*:/g, '$1"$2":');
				var json = JSON.parse(fixed);
				return syntaxHighlightStats(JSON.stringify(json, null, 2));
			} catch(e2) {
				// Still not valid, return as-is
				return text;
			}
		}
	}
</script>
<style>
	.stats-json-btn {
		background: #f0f0f0;
		border: 1px solid #ccc;
		border-radius: 3px;
		padding: 2px 6px;
		cursor: pointer;
		font-family: monospace;
		font-size: 0.85em;
		color: #666;
		margin-right: 5px;
	}
	.stats-json-btn:hover {
		background: #e0e0e0;
	}
	.stats-json-btn.active {
		background: #d0e8d0;
		border-color: #7BAF9B;
	}
	.stats-json-row {
		display: none;
	}
	.stats-json-content {
		background: #1e1e1e;
		color: #d4d4d4;
		padding: 12px;
		border-radius: 4px;
		font-family: 'Consolas', 'Monaco', monospace;
		font-size: 0.9em;
		white-space: pre-wrap;
		word-break: break-all;
		margin: 5px 0;
		max-height: 400px;
		overflow: auto;
	}
</style>`, orderBy)
	asc := "<i class='fa fa-sort-asc'/>"
	desc := "<i class='fa fa-sort-desc'/>"
	html += `<div align='left'>`
	if download == "" {
		html += `
<!-- Header Bar -->
<div style='display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;'>
	<h2 style='margin: 0; color: #444; font-size: 1.4em;'><i class='fa fa-info-circle' style='color: #1565c0;'></i> Slow Query Patterns</h2>
	<button id="download" onClick="downloadStats(); return false;"
		class="download-btn"><i class="fa fa-download"></i> Download</button>
</div>`
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
		html += fmt.Sprintf(`<th valign='middle'>index <label title='Show COLLSCAN only' style='cursor: pointer; font-weight: normal; font-size: 0.85em;'><input type='checkbox' id='collscan' onchange='getSlowopsStats(); return false;' %v> only</label></th>`, checked)
	} else {
		html += "<th valign='middle'>index</th>"
	}
	html += `<th>query pattern</th><th style='width: 40px;'></th>
		</tr>
{{$merge := .Merge}}
{{range $n, $value := .Ops}}
		<tr>
		{{if $merge}}
			<td align='right'>{{ add $n 1 }} {{getMarkerHTML $value.Marker}}</td>
		{{else}}
			<td align='right'>{{ add $n 1 }}</td>
		{{end}}
			<td style='white-space: nowrap;'>{{ $value.Op }}</td>
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
			<td align='center'><button id='btn-stats-{{$n}}' class='stats-json-btn' onclick='toggleStatsJson({{$n}})' title='View formatted'>{}</button></td>
		</tr>
		<tr id='json-stats-{{$n}}' class='stats-json-row'>
			<td colspan='11' style='padding: 5px 10px;'>
				<div style='display: flex; gap: 20px;'>
					<div style='flex: 1;'>
						<div style='font-weight: bold; margin-bottom: 5px; color: #666;'>Index:</div>
						<pre class='stats-json-content' id='index-content-{{$n}}'>{{ $value.Index }}</pre>
					</div>
					<div style='flex: 2;'>
						<div style='font-weight: bold; margin-bottom: 5px; color: #666;'>Query Pattern:</div>
						<pre class='stats-json-content' id='pattern-content-{{$n}}'>{{ $value.QueryPattern }}</pre>
					</div>
				</div>
			</td>
		</tr>
{{end}}
	</table>
	</div>
	<div align='center'><hr/><p/>{{.Version}}</div>
</div>
`
	return html
}
