/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * audit_template.go
 */

package hatchet

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/simagix/gox"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetAuditTablesTemplate returns HTML
func GetAuditTablesTemplate(download string) (*template.Template, error) {
	html := headers
	if download == "" {
		html += getContentHTML()
	}
	html += `{{$name := .Hatchet}}
<script>
	function downloadAudit() {
		anchor = document.createElement('a');
		anchor.download = '{{.Hatchet}}_audit.html';
		anchor.href = '/hatchets/{{.Hatchet}}/stats/audit?download=true';
		anchor.dataset.downloadurl = ['text/html', anchor.download, anchor.href].join(':');
		anchor.click();
	}
</script>`
	if download == "" {
		html += `
<!-- Header Bar -->
<div style='display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;'>
	<h2 style='margin: 0; color: #444; font-size: 1.4em;'><i class='fa fa-shield' style='color: #2e7d32;'></i> Audit Report</h2>
	<button id="download" onClick="downloadAudit(); return false;"
		class="download-btn"><i class="fa fa-download"></i> Download</button>
</div>`
	} else {
		html += "<div align='center'>{{.Summary}}</div>"
	}
	html += `
<!-- Quick Stats Bar -->
<div style='display: flex; gap: 20px; margin: 15px 5px; flex-wrap: wrap;'>
	<div style='background: linear-gradient(135deg, #e8f5e9 0%, #c8e6c9 100%); padding: 15px 25px; border-radius: 12px; flex: 1; min-width: 180px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);'>
		<div style='font-size: 0.85em; color: #666; text-transform: uppercase; letter-spacing: 0.5px;'>Total Slow Ops</div>
		<div style='font-size: 1.8em; font-weight: bold; color: #2e7d32;'>{{getTotalOps .Data}}</div>
	</div>
	<div style='background: linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%); padding: 15px 25px; border-radius: 12px; flex: 1; min-width: 180px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);'>
		<div style='font-size: 0.85em; color: #666; text-transform: uppercase; letter-spacing: 0.5px;'>Time Range</div>
		<div style='font-size: 1.1em; font-weight: bold; color: #1565c0;'>{{getTimeRange .Info}}</div>
	</div>
	<div style='background: linear-gradient(135deg, #fff3e0 0%, #ffe0b2 100%); padding: 15px 25px; border-radius: 12px; flex: 1; min-width: 180px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);'>
		<div style='font-size: 0.85em; color: #666; text-transform: uppercase; letter-spacing: 0.5px;'>Unique Clients</div>
		<div style='font-size: 1.8em; font-weight: bold; color: #ef6c00;'>{{getUniqueClients .Data}}</div>
	</div>
	{{if hasCollscan .Data}}
	<div style='background: linear-gradient(135deg, #ffebee 0%, #ffcdd2 100%); padding: 15px 25px; border-radius: 12px; flex: 1; min-width: 180px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);'>
		<div style='font-size: 0.85em; color: #666; text-transform: uppercase; letter-spacing: 0.5px;'>COLLSCAN Ops</div>
		<div style='font-size: 1.8em; font-weight: bold; color: #c62828;'>{{getCollscanCount .Data}}</div>
	</div>
	{{end}}
</div>

<div style='margin: 5px 5px; width=100%; clear: left;'>
	  <table style='border: none; margin: 10px 10px; width=100%; clear: left;' width='100%'>
		{{$flag := coinToss}}
		<tr><td style='border:none; vertical-align: top; padding: 5px; background-color: #F3F7F4;'>
			<img class='rotate23' src='data:image/png;base64,{{ assignConsultant $flag }}'></img></td>
			<td class='summary'>
				{{getInfoSummary .Info $flag}}<p/>{{getStatsSummary .Data}}</td>
		</tr>
	  </table>
	</div>
<!-- Security & Warnings Section -->
{{if or (hasData .Data "exception") (hasData .Data "driver") (hasData .Data "failed")}}
<div style='clear: both; height: 30px;'></div>
<h3 style='margin: 10px 10px 10px 10px; color: #555; border-bottom: 2px solid #ddd; padding-bottom: 8px;'>
	<i class='fa fa-shield' style='color: #c62828;'></i> Security & Warnings
</h3>
{{end}}

{{if hasData .Data "exception"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><button class='btn'
			onClick="javascript:loadData('/hatchets/{{.Hatchet}}/logs/all?severity=W'); return false;">
			<i class='fa fa-search'></i></button>Exceptions</caption>
		<tr><th></th><th>Severity</th><th>Total</th></tr>
	{{range $n, $val := index .Data "exception"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:loadData('/hatchets/{{$name}}/logs/all?severity={{slice $val.Name 0 1}}'); return false;"><i class='fa fa-search'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "driver"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><button class='btn'><i class='fa fa-comment-o'></i></button>Drivers Compatibility</caption>
		<tr><th></th><th>Driver</th><th>Version</th><th>IP</th><th>Compatibility</th></tr>
	{{$mver := .Info.Version}}
	{{range $n, $val := index .Data "driver"}}
		<tr><td align=right>{{add $n 1}}</td>
			<td>{{index $val.Values 0}}</td><td>{{index $val.Values 1}}</td>
			<td>{{$val.Name}}</td>
			{{$err := checkDriver $mver $val.Values}}
			{{if eq $err nil}}
				<td align='center'><i class='fa fa-check'></i></td>
			{{else}}
				<td><mark>{{$err}}</mark></td>
			{{end}}
		</tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "failed"}}
	<table style='float: left; margin: 10px 10px; clear: left;'>
		<caption><button class='btn'
			onClick="javascript:loadData('/hatchets/{{.Hatchet}}/logs/all?context=failed'); return false;">
			<i class='fa fa-search'></i></button>Failed Operations</caption>
		<tr><th></th><th>Failed Operation</th><th>Total</th></tr>
	{{range $n, $val := index .Data "failed"}}
		<tr><td align=right>{{add $n 1}}</td>
			<td>
				<button class='btn' onClick="javascript:loadData('/hatchets/{{$name}}/logs/all?context={{$val.Name}}'); return false;"><i class='fa fa-search'></i></button>{{$val.Name}}
			</td>
			<td align=right>{{getFormattedNumber $val.Values 0}}</td>
		</tr>
	{{end}}
	</table>
{{end}}

<!-- Connections & Clients Section -->
{{if or (hasData .Data "ip") (hasData .Data "duration")}}
<div style='clear: both; height: 30px;'></div>
<h3 style='margin: 10px 10px 10px 10px; color: #555; border-bottom: 2px solid #ddd; padding-bottom: 8px;'>
	<i class='fa fa-plug' style='color: #1565c0;'></i> Connections & Clients
</h3>
{{end}}

{{if hasData .Data "ip"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><button class='btn'
			onClick="javascript:loadData('/hatchets/{{.Hatchet}}/charts/connections?type=accepted'); return false;">
			<i class='fa fa-pie-chart'></i></button>Stats by IPs</caption>
		<tr><th></th><th>IP</th><th>Accepted Connections</th>{{if ge (len (index (index .Data "ip") 0).Values) 3}}<th>Closed Connections</th>{{end}}<th>Response Length</th></tr>
	{{range $n, $val := index .Data "ip"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:loadData('/hatchets/{{$name}}/charts/reslen-ip?ip={{$val.Name}}'); return false;"><i class='fa fa-pie-chart'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td>{{if ge (len $val.Values) 3}}<td align=right>{{getFormattedNumber $val.Values 2}}</td>{{end}}<td align=right>{{getFormattedSize $val.Values 1}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "duration"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><span style="font-size: 16px; padding: 5px 5px;"><i class="fa fa-clock-o"></i></span>Long Lasting Connections</caption>
		<tr><th></th><th>Context</th><th>Duration</th></tr>
	{{range $n, $val := index .Data "duration"}}
			<tr><td align=right>{{add $n 1}}</td>
				<td><button class='btn' onClick="javascript:loadData('/hatchets/{{$name}}/logs/all?context={{getContext $val.Name}}'); return false;">
					<i class='fa fa-search'></i></button>{{$val.Name}}
				</td>
				<td align=right>{{getFormattedDuration $val.Values 0}}</td>
			</tr>
	{{end}}
	</table>
{{end}}

<!-- Performance Section -->
{{if or (hasData .Data "op") (hasData .Data "ns") (hasData .Data "appname")}}
<div style='clear: both; height: 30px;'></div>
<h3 style='margin: 10px 10px 10px 10px; color: #555; border-bottom: 2px solid #ddd; padding-bottom: 8px;'>
	<i class='fa fa-tachometer' style='color: #ef6c00;'></i> Performance
	<button class='btn' style='margin-left: 10px;' onClick="javascript:loadData('/hatchets/{{.Hatchet}}/stats/slowops'); return false;" title='View detailed slow query patterns'>
		<i class='fa fa-list'></i> View Stats
	</button>
</h3>
{{end}}

{{if hasData .Data "op"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><button class='btn'
			onClick="javascript:loadData('/hatchets/{{.Hatchet}}/charts/ops?type=stats'); return false;">
			<i class='fa fa-area-chart'></i></button>Operations Stats</caption>
		<tr><th></th><th>Operation</th><th>Total</th></tr>
	{{range $n, $val := index .Data "op"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:loadData('/hatchets/{{$name}}/charts/ops?type=stats&op={{$val.Name}}'); return false;"><i class='fa fa-area-chart'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "ns"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><button class='btn'
			onClick="javascript:loadData('/hatchets/{{.Hatchet}}/charts/reslen-ns?ns='); return false;">
			<i class='fa fa-pie-chart'></i></button>Stats by Namespaces</caption>
		<tr><th></th><th>Namespace</th><th>Accessed</th><th>Response Length</th></tr>
	{{range $n, $val := index .Data "ns"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:loadData('/hatchets/{{$name}}/logs/all?context={{$val.Name}}'); return false;"><i class='fa fa-search'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td><td align=right>{{getFormattedSize $val.Values 1}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "appname"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><button class='btn'
			onClick="javascript:loadData('/hatchets/{{.Hatchet}}/charts/reslen-appname?appname='); return false;">
			<i class='fa fa-pie-chart'></i></button>Stats by AppName</caption>
		<tr><th></th><th>AppName</th><th>Accessed</th><th>Response Length</th></tr>
	{{range $n, $val := index .Data "appname"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:loadData('/hatchets/{{$name}}/logs/all?context={{$val.Name}}'); return false;"><i class='fa fa-search'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td><td align=right>{{getFormattedSize $val.Values 1}}</td></tr>
	{{end}}
	</table>
{{end}}
	<div style='clear: left;' align='center'><hr/><p/>{{.Version}}</div>
`
	if download == "" {
		html += "</div><!-- end content-container -->"
	}
	html += "</body></html>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"hasData": func(data map[string][]NameValues, key string) bool {
			return len(data[key]) > 0
		},
		"hasCollscan": func(data map[string][]NameValues) bool {
			for _, item := range data["collscan"] {
				if item.Name == "count" && len(item.Values) > 0 {
					if count, ok := item.Values[0].(int); ok && count > 0 {
						return true
					}
				}
			}
			return false
		},
		"getTotalOps": func(data map[string][]NameValues) string {
			printer := message.NewPrinter(language.English)
			total := 0
			for _, item := range data["op"] {
				if len(item.Values) > 0 {
					if count, ok := item.Values[0].(int); ok {
						total += count
					}
				}
			}
			return printer.Sprintf("%d", total)
		},
		"getTimeRange": func(info HatchetInfo) string {
			if info.Start == "" || info.End == "" {
				return "N/A"
			}
			layout := "2006-01-02T15:04:05"
			startTime := info.Start
			if len(startTime) > 19 {
				startTime = startTime[:19]
			}
			endTime := info.End
			if len(endTime) > 19 {
				endTime = endTime[:19]
			}
			stime, err1 := time.Parse(layout, startTime)
			etime, err2 := time.Parse(layout, endTime)
			if err1 != nil || err2 != nil {
				return "N/A"
			}
			seconds := etime.Unix() - stime.Unix()
			return gox.GetDurationFromSeconds(float64(seconds))
		},
		"getUniqueClients": func(data map[string][]NameValues) string {
			return fmt.Sprintf("%d", len(data["ip"]))
		},
		"getCollscanCount": func(data map[string][]NameValues) string {
			printer := message.NewPrinter(language.English)
			for _, item := range data["collscan"] {
				if item.Name == "count" && len(item.Values) > 0 {
					return printer.Sprintf("%v", item.Values[0])
				}
			}
			return "0"
		},
		"numPrinter": func(n interface{}) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", ToInt(n))
		},
		"getContext": func(s string) string {
			toks := strings.Split(s, " ")
			if len(toks) == 0 {
				return s
			}
			return toks[0]
		},
		"assignConsultant": func(sage bool) string {
			if sage {
				return SAGE_PNG
			}
			return SIMONE_PNG
		},
		"checkDriver": func(version string, values []interface{}) error {
			return CheckDriverCompatibility(version, values[0].(string), values[1].(string))
		},
		"getFormattedNumber": func(numbers []interface{}, i int) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", numbers[i])
		},
		"coinToss": func() bool {
			// rand.Seed(time.Now().UnixNano())
			// randomNum := rand.Intn(2)
			// return (randomNum%2 == 0)
			return false // always Simone
		},
		"getDurationFromSeconds": func(s int) string {
			return gox.GetDurationFromSeconds(float64(s))
		},
		"getFormattedDuration": func(numbers []interface{}, i int) string {
			return gox.GetDurationFromSeconds(float64(numbers[i].(int)))
		},
		"getStorageSize": func(s int) string {
			return gox.GetStorageSize(float64(s))
		},
		"getFormattedSize": func(numbers []interface{}, i int) string {
			return gox.GetStorageSize(numbers[i])
		},
		"getInfoSummary": func(info HatchetInfo, sage bool) template.HTML {
			var html = "Hey there! My name is <i>Simone</i> and here is the summary I've prepared for you. "
			if sage {
				html = "Hello, my name is <i>Sage</i> and I'd like to share my thoughts with you on the findings. "
			}
			if info.Version == "" {
				html += "There is not enough information in the log to determine what MongoDB version is used."
			} else {
				html += fmt.Sprintf("So, the server running MongoDB is using the <span style='color: orange;'>%v</span> edition of version <span style='color: orange;'>%v</span>", info.Module, info.Version)
				if strings.Compare(MIN_MONGO_VER, info.Version) == 1 {
					html += ", which is kinda old now. <mark>It's probably a good idea to upgrade to a more recent version of MongoDB soon</mark>. "
				} else {
					html += ". "
				}
				if info.Arch != "" && info.OS != "" {
					html += fmt.Sprintf("The server is on a <span style='color: orange;'>%v</span> architecture server that running the <span style='color: orange;'>%v</span> operating system. ", info.Arch, info.OS)
				}
				if info.Module != "" && info.Module != "enterprise" {
					html += fmt.Sprintf("Although %v edition works, it is recommended to upgrade to the enterprise edition or migrate to Atlas. ", info.Module)
				}
				if info.Start != "" && info.End != "" {
					layout := "2006-01-02T15:04:05"
					startTime := info.Start[:19]
					endTime := info.End[:19]
					stime, _ := time.Parse(layout, startTime)
					etime, _ := time.Parse(layout, endTime)
					seconds := etime.Unix() - stime.Unix()
					days := gox.GetDurationFromSeconds(float64(seconds))
					if days == " 1.0 days" {
						days = "1 day"
					}
					html += fmt.Sprintf("The first log was dated on <span style='color: orange;'>%v</span>, it ran for %v, ending on <span style='color: orange;'>%v</span>. ",
						stime.Format(time.ANSIC), days, etime.Format(time.ANSIC))
					if seconds > (24*60*60 + 5*60) {
						html += fmt.Sprintf("<mark>The duration of %v is greater than a day for a single log file, and rotating logs regularly is recommended.</mark> ",
							gox.GetDurationFromSeconds(float64(seconds)))
					}
				}
				if info.Provider != "" && info.Region != "" {
					html += fmt.Sprintf("Bravo for the decision to host your servers on Atlas <span style='color: orange;'>%s</span> in the <span style='color: orange;'>%s</span> region. ",
						info.Provider, info.Region)
					html += "A gentle reminder to <mark>ensure your application servers reside in the same region</mark> which can save you more than a few bucks from data transfer charges. "
				}
				if len(info.Drivers) == 1 {
					for _, driver := range info.Drivers {
						for key, value := range driver {
							html += "For the application driver, I found a driver information and it was "
							html += fmt.Sprintf("<span style='color: orange;'>%s</span> version <span style='color: orange;'>%s</span>. ", key, value)
						}
					}
					html += "See Drivers Compatibility table below for a list of drivers in use. "
				} else if len(info.Drivers) > 1 {
					html += "It looks like your applications have used a number of drivers, and they are "
					cnt := 0
					for _, driver := range info.Drivers {
						for key, value := range driver {
							cnt++
							if cnt == len(info.Drivers) {
								html += fmt.Sprintf("and <span style='color: orange;'>%s</span> version <span style='color: orange;'>%s</span>. ", key, value)
							} else {
								html += fmt.Sprintf("<span style='color: orange;'>%s</span> version <span style='color: orange;'>%s</span>, ", key, value)
							}
						}
					}
					html += "See Drivers Compatibility table below for a list of drivers in use. "
				}
			}
			return template.HTML(html)
		},
		"getStatsSummary": func(data map[string][]NameValues) template.HTML {
			var html string
			printer := message.NewPrinter(language.English)
			var totalImpact float64
			for key, docs := range data {
				if key == "exception" && len(docs) > 0 {
					html += printer.Sprintf("Hmmm, I found <span style='color: orange;'>%d</span> warning (or more severe) ", len(docs))
					if len(docs) < 2 {
						html += "message. "
					} else {
						html += "messages. "
					}
				} else if key == "ip" && len(docs) > 0 {
					conns := 0
					for _, doc := range docs {
						conns += doc.Values[0].(int)
					}
					html += printer.Sprintf("During the time, there were a total of <span style='color: orange;'>%d</span> accepted connections from <span style='color: orange;'>%d</span> ", conns, len(docs))
					if len(docs) < 2 {
						html += "client. "
					} else {
						html += "different clients. "
					}
				} else if key == "ns" && len(docs) > 0 {
					count := 0
					for _, doc := range docs {
						count += doc.Values[0].(int)
					}
					html += printer.Sprintf("As many as <span style='color: orange;'>%d</span> different namespaces were accessed a total of <span style='color: orange;'>%d</span> times. ", len(docs), count)
					reslen := 0
					for _, doc := range docs {
						reslen += doc.Values[1].(int)
					}
					html += printer.Sprintf("The total response length was around <span style='color: orange;'>%v</span>. ", gox.GetStorageSize(reslen))
				} else if key == "stats" && len(docs) > 0 {
					for _, doc := range docs {
						if doc.Name == "maxConns" {
							milli := doc.Values[0].(int)
							if milli == 0 {
								html += "I didn't find any connections information. "
								continue
							}
							html += printer.Sprintf("At one point, the number of opened connections reached <span style='color: orange;'>%d</span>", doc.Values[0])
							if milli > 1000 {
								mem := milli * (1024 * 1024)
								html += printer.Sprintf(", <mark>which could take up about %s of memory</mark>. ", gox.GetStorageSize(float64(mem)))
							} else {
								html += ". "
							}
						} else if doc.Name == "maxMilli" && doc.Values[0].(int) > 0 {
							milli := doc.Values[0].(int)
							html += "The slowest operation took "
							if milli < 1000 {
								html += printer.Sprintf("<span style='color: orange;'>%d</span> milliseconds. ", doc.Values[0])
							} else {
								seconds := float64(milli) / 1000
								html += printer.Sprintf("<span style='color: orange;'>%s</span>. ", gox.GetDurationFromSeconds(seconds))
							}
						} else if doc.Name == "avgMilli" && doc.Values[0].(int) > 0 {
							milli := doc.Values[0].(int)
							html += printer.Sprintf("Moreover, the average operation time was <span style='color: orange;'>%d</span> milliseconds", milli)
							if milli > 100 {
								html += `, where operation time <mark>greater than 100 milliseconds is, IMO, "slow"</mark>. `
							} else {
								html += ". "
							}
						} else if doc.Name == "totalMilli" && doc.Values[0].(int) > 0 {
							seconds := float64(doc.Values[0].(int)) / 1000
							if seconds < (10 * 60) { // should be calculated with duration
								html += printer.Sprintf("The total impact time from slowest operations was %s. ", gox.GetDurationFromSeconds(seconds))
							} else if seconds < (60 * 60) {
								html += printer.Sprintf("The total impact time from slowest operations was, ouch, <span style='color: orange;'>%s</span>. ", gox.GetDurationFromSeconds(seconds))
							} else {
								totalImpact = seconds
							}
						}
					}
				} else if key == "collscan" && len(docs) > 0 {
					html += "Let's move to the performance evaluation. "
					for _, doc := range docs {
						if doc.Name == "count" {
							html += printer.Sprintf(`I found <span style='color: orange;'>%d</span> with <mark><i>COLLSCAN</i> </mark>plan summary. `, doc.Values[0])
						} else if doc.Name == "totalMilli" {
							seconds := float64(doc.Values[0].(int)) / 1000
							html += printer.Sprintf(`The <i>COLLSCAN</i> caused a total of <span style='color: orange;'>%s</span> wasted. `, gox.GetDurationFromSeconds(seconds))
						}
					}
				}
			}
			if totalImpact > 0 {
				html += "OMG, the total impact from slowest operations was "
				html += printer.Sprintf("<span style='color: orange;'>%s</span>, this may be a problem of lacking resources.  Please review the sizing training videos below. ", gox.GetDurationFromSeconds(totalImpact))
				html += "<div style='padding: 10px'>"
				html += `<iframe width="400" height="225" src="https://www.youtube.com/embed/kObLsYJAruI" title="Survery Your MongoDB Land" style="margin-right: 5px;” frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>`
				html += `<iframe width="400" height="225" src="https://www.youtube.com/embed/equz1z0igv0" title="Bond - MongoDB Sharded Cluster Analysis" style="margin-right: 5px;” frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>`
				html += `<iframe width="400" height="225" src="https://www.youtube.com/embed/0AAMw_q1E4o" title="Sizing a MongoDB Cluster 1" style="margin-right: 5px;” frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>`
				html += `<iframe width="400" height="225" src="https://www.youtube.com/embed/n1wORkr_1xI" title="Sizing a MongoDB Cluster 2" style="margin-right: 5px;” frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>`
				html += "</div>"
			}
			if len(data) > 0 {
				html += "<p/>Check out the details below for additional information about this server."
			}
			return template.HTML(html)
		}}).Parse(html)
}
