// Copyright 2022-present Kuei-chun Chen. All rights reserved.
// logs_template.go
package hatchet

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

var (
	SEVERITIES = []string{"F", "E", "W", "I", "D", "D2"}
	SEVERITY_M = map[string]string{"F": "FATAL", "E": "ERROR", "W": "WARN", "I": "INFO",
		"D": "DEBUG", "D2": "DEBUG2"}
)

// GetLogTableTemplate returns HTML
func GetLogTableTemplate(attr string, download string) (*template.Template, error) {
	html := headers
	if download == "" {
		html += getContentHTML()
	}
	if attr == "slowops" {
		html += getSlowOpsLogsTable(download)
	} else {
		html += getLegacyLogsTable()
	}
	if download == "" {
		html += "</div><!-- end content-container -->"
	}
	html += "</body></html>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"getComponentOptions": func(item string) template.HTML {
			arr := []string{}
			comps := []string{"ACCESS", "ASIO", "COMMAND", "CONNPOOL", "CONTROL", "ELECTION", "FTDC", "INDEX", "INITSYNC", "NETWORK",
				"QUERY", "RECOVERY", "REPL", "SHARDING", "STORAGE", "WRITE"}
			for _, v := range comps {
				selected := ""
				if v == item {
					selected = "SELECTED"
				}
				arr = append(arr, fmt.Sprintf("<option value='%v' %v>%v</option>", v, selected, v))
			}
			return template.HTML(strings.Join(arr, "\n"))
		},
		"getSeverityOptions": func(item string) template.HTML {
			arr := []string{}
			for _, v := range SEVERITIES {
				selected := ""
				if v == item {
					selected = "SELECTED"
				}
				arr = append(arr, fmt.Sprintf("<option value='%v' %v>%v</option>", v, selected, SEVERITY_M[v]))
			}
			return template.HTML(strings.Join(arr, "\n"))
		},
		"getMarkerHTML": func(marker int) template.HTML {
			return template.HTML(GetMarkerHTML(marker))
		},
		"highlightLog": func(log string, params ...string) template.HTML {
			return template.HTML(highlightLog(log, params...))
		},
		"formatDateTime": func(str string) string {
			return strings.Replace(str, "T", " ", 1)
		}}).Parse(html)
}

func highlightLog(log string, params ...string) string {
	re := regexp.MustCompile(`("?(planSummary)"?:\s?"(.*?)")`)
	log = re.ReplaceAllString(log, "<mark>$1</mark>")
	re = regexp.MustCompile(`((\d+ms$))`)
	log = re.ReplaceAllString(log, "<mark>$1</mark>")
	re = regexp.MustCompile(`(("?(keysExamined|keysInserted|docsExamined|nreturned|nMatched|nModified|ndeleted|ninserted|reslen|durationMillis)"?:)\d+)`)
	log = re.ReplaceAllString(log, "<mark>$1</mark>")
	re = regexp.MustCompile(`(?i)("?(errMsg)"?:\s?"(.*?)"|planSummary:\s?"?COLLSCAN"?)`)
	log = re.ReplaceAllString(log, "<span style='color: red; font-weight: bold;'>$1</span>")
	for _, param := range params {
		if param != "" {
			re = regexp.MustCompile("(?i)(" + param + `(:\s?\".*?\")?)`)
			log = re.ReplaceAllString(log, "<mark>$1</mark>")
		}
	}
	return log
}

func getSlowOpsLogsTable(download string) string {
	template := `
<script>
	function downloadTopN() {
		anchor = document.createElement('a');
		anchor.download = '{{.Hatchet}}_topn.html';
		anchor.href = '/hatchets/{{.Hatchet}}/logs/slowops?download=true';
		anchor.dataset.downloadurl = ['text/html', anchor.download, anchor.href].join(':');
		anchor.click();
	}
</script>`
	if download == "" {
		template += `
<!-- Header Bar -->
<div style='display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;'>
	<h2 style='margin: 0; color: #444; font-size: 1.4em;'><i class='fa fa-list-ol' style='color: #ef6c00;'></i> Top N Slowest Operations</h2>
	<button id="download" onClick="downloadTopN(); return false;"
		class="download-btn"><i class="fa fa-download"></i> Download</button>
</div>`
	} else {
		template += `<div align='center'>{{.Summary}}</div>`
	}
	template += `
<div align='center'>
	<table width='100%'>
		<tr>
			<th>#</th>
			<th style='width: 40px;'></th>
		{{ if .Merge }}
			<th>M</th>
		{{ end }}
			<th>date</th>
			<th>S</th>
			<th>component</th>
			<th>context</th>
			<th>message</th>
		</tr>
{{$hatchet := .Hatchet}}
{{$merge := .Merge}}
{{range $n, $value := .Logs}}
		<tr>
			<td align='right'>{{ add $n 1 }}</td>
			<td align='center'><button id='btn-topn-{{$n}}' class='json-toggle-btn' onclick='toggleJsonView("topn-{{$n}}")' title='View formatted JSON'>{}</button></td>
		{{ if $merge }}
			<td>{{ getMarkerHTML $value.Marker }}</td>
		{{ end }}
			<td>{{ formatDateTime $value.Timestamp }}</td>
			<td>{{ $value.Severity }}</td>
			<td>{{ $value.Component }}</td>
			<td><a href='/hatchets/{{$hatchet}}/logs/all?context={{$value.Context}}'>{{ $value.Context }}</a></td>
			<td class='break'>{{ highlightLog $value.Message }}</td>
		</tr>
		<tr id='json-topn-{{$n}}' class='json-row'>
			<td colspan='{{ if $merge }}8{{ else }}7{{ end }}' style='padding: 0 10px;'>
				<pre class='json-content' id='json-content-topn-{{$n}}'>{{ $value.Message }}</pre>
			</td>
		</tr>
{{end}}
	</table>
	<div style='clear: left;' align='center'><hr/><p/>{{.Version}}</div>
</div>
<script>
document.addEventListener('DOMContentLoaded', function() {
	// Format all JSON content on page load
	document.querySelectorAll('.json-content').forEach(function(el) {
		try {
			var json = JSON.parse(el.textContent);
			el.innerHTML = syntaxHighlight(JSON.stringify(json, null, 2));
		} catch(e) {
			// Not valid JSON, leave as-is
		}
	});
});

function syntaxHighlight(json) {
	json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
	return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
		var cls = 'json-number';
		if (/^"/.test(match)) {
			if (/:$/.test(match)) {
				cls = 'json-key';
				// Highlight important keys
				if (/planSummary|errMsg|durationMillis/i.test(match)) {
					cls = 'json-key json-highlight';
				}
			} else {
				cls = 'json-string';
				// Highlight COLLSCAN
				if (/COLLSCAN/.test(match)) {
					cls = 'json-error';
				}
			}
		} else if (/true|false/.test(match)) {
			cls = 'json-boolean';
		} else if (/null/.test(match)) {
			cls = 'json-null';
		}
		return '<span class="' + cls + '">' + match + '</span>';
	});
}
</script>
`
	return template
}

func getLegacyLogsTable() string {
	template := `
<!-- Header Bar -->
<div style='display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;'>
	<h2 style='margin: 0; color: #444; font-size: 1.4em;'><i class='fa fa-search' style='color: #7b1fa2;'></i> Search Logs</h2>
</div>
  <div style="float: left; margin-right: 20px; clear: left;">
	<label><i class="fa fa-leaf"></i></label>
	<select id='component'>
		<option value=''>select a component</option>
		{{getComponentOptions .Component}}
	</select>
  </div>

  <div style="float: left; margin-right: 20px;">
	<label><i class="fa fa-exclamation"></i></label>
	<select id='severity'>
		<option value=''>select a severity</option>
		{{getSeverityOptions .Severity}}
	</select>
  </div>

  <div style="float: left; margin-right: 20px;">
	<label><i class="fa fa-search"></i></label>
	<input id='context' type='text' value='{{.Context}}' size='30'/>
	<button id="find" onClick="findLogs()" class="button" style="float: right;">Find</button>
  </div>

<div style="clear: both; padding-top: 25px;">
{{ if .Logs }}
	<div style="margin-bottom: 5px;">
		<span style="font-weight: bold;">Total records: {{.TotalCount}}</span>
	</div>
	{{if .HasMore}}
		<button onClick="javascript:loadData('{{.URL}}'); return false;"
			class="btn" style="float: right; clear: right"><i class="fa fa-arrow-right"></i></button>
	{{end}}
	<table width='100%'>
		<tr>
			<th>#</th>
			<th style='width: 40px;'></th>
	{{ if .Merge }}
			<th>M</th>
	{{end}}
			<th>date</th>
			<th>S</th>
			<th>component</th>
			<th>context</th>
			<th>message</th>
		</tr>
	{{$search := .Context}}
	{{$seq := .Seq}}
	{{$hatchet := .Hatchet}}
	{{$merge := .Merge}}
	{{range $n, $value := .Logs}}
		<tr>
			<td align='right'>{{ add $n $seq }}</td>
			<td align='center'><button id='btn-search-{{$n}}' class='json-toggle-btn' onclick='toggleJsonView("search-{{$n}}")' title='View formatted JSON'>{}</button></td>
		{{ if $merge }}
			<td>{{ getMarkerHTML $value.Marker }}</td>
		{{ end }}
			<td>{{ formatDateTime $value.Timestamp }}</td>
			<td>{{ $value.Severity }}</td>
			<td>{{ $value.Component }}</td>
			<td><a href='/hatchets/{{$hatchet}}/logs/all?context={{$value.Context}}'>{{ $value.Context }}</a></td>
			<td class='break'>{{ highlightLog $value.Message $search }}</td>
		</tr>
		<tr id='json-search-{{$n}}' class='json-row'>
			<td colspan='{{ if $merge }}8{{ else }}7{{ end }}' style='padding: 0 10px;'>
				<pre class='json-content' id='json-content-search-{{$n}}'>{{ $value.Message }}</pre>
			</td>
		</tr>
	{{end}}
	</table>
	{{if .HasMore}}
		<button onClick="javascript:loadData('{{.URL}}'); return false;"
			class="btn" style="float: right; clear: right;"><i class="fa fa-arrow-right"></i></button>
	{{end}}
<div align='center'><hr/><p/>{{.Version}}</div>
{{end}}
</div>
<script>
	var input = document.getElementById("context");
	input.addEventListener("keypress", function(event) {
		if (event.key === "Enter") {
			event.preventDefault();
			document.getElementById("find").click();
		}
	});

	function findLogs() {
		var sel = document.getElementById('component')
		var component = sel.options[sel.selectedIndex].value;
		sel = document.getElementById('severity')
		var severity = sel.options[sel.selectedIndex].value;
		var context = document.getElementById('context').value
		loadData('/hatchets/{{.Hatchet}}/logs/all?component='+component+'&severity='+severity+'&context='+context);
	}

	// Format all JSON content on page load
	document.querySelectorAll('.json-content').forEach(function(el) {
		try {
			var json = JSON.parse(el.textContent);
			el.innerHTML = syntaxHighlight(JSON.stringify(json, null, 2));
		} catch(e) {
			// Not valid JSON, leave as-is
		}
	});

	function syntaxHighlight(json) {
		json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
		return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
			var cls = 'json-number';
			if (/^"/.test(match)) {
				if (/:$/.test(match)) {
					cls = 'json-key';
					// Highlight important keys
					if (/planSummary|errMsg|durationMillis/i.test(match)) {
						cls = 'json-key json-highlight';
					}
				} else {
					cls = 'json-string';
					// Highlight COLLSCAN
					if (/COLLSCAN/.test(match)) {
						cls = 'json-error';
					}
				}
			} else if (/true|false/.test(match)) {
				cls = 'json-boolean';
			} else if (/null/.test(match)) {
				cls = 'json-null';
			}
			return '<span class="' + cls + '">' + match + '</span>';
		});
	}
</script>
`
	return template
}
