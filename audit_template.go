/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * audit_template.go
 */

package hatchet

import (
	"fmt"
	"html/template"
	"math/rand"
	"strings"
	"time"

	"github.com/simagix/gox"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetAuditTablesTemplate returns HTML
func GetAuditTablesTemplate() (*template.Template, error) {
	html := headers + getContentHTML()
	html += `{{$name := .Hatchet}}
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
{{if hasData .Data "exception"}}
	<table style='float: left; margin: 10px 10px; clear: left;'>
		<caption><button class='btn'
			onClick="javascript:location.href='/hatchets/{{.Hatchet}}/logs/all?severity=W'; return false;">
			<i class='fa fa-search'></i></button>Exceptions</caption>
		<tr><th></th><th>Severity</th><th>Total</th></tr>
	{{range $n, $val := index .Data "exception"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:location.href='/hatchets/{{$name}}/logs/all?severity={{slice $val.Name 0 1}}'; return false;"><i class='fa fa-search'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "failed"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><button class='btn'
			onClick="javascript:location.href='/hatchets/{{.Hatchet}}/logs/all?context=failed'; return false;">
			<i class='fa fa-search'></i></button>Failed Operations</caption>
		<tr><th></th><th>Failed Operation</th><th>Total</th></tr>
	{{range $n, $val := index .Data "failed"}}
		<tr><td align=right>{{add $n 1}}</td>
			<td>
				<button class='btn' onClick="javascript:location.href='/hatchets/{{$name}}/logs/all?context={{$val.Name}}'; return false;"><i class='fa fa-search'></i></button>{{$val.Name}}
			</td>
			<td align=right>{{getFormattedNumber $val.Values 0}}</td>
		</tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "op"}}
	<table style='float: left; margin: 10px 10px; clear: left;'>
		<caption><button class='btn'
			onClick="javascript:location.href='/hatchets/{{.Hatchet}}/charts/ops?type=stats'; return false;">
			<i class='fa fa-area-chart'></i></button>Operations Stats</caption>
		<tr><th></th><th>Operation</th><th>Total</th></tr>
	{{range $n, $val := index .Data "op"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:location.href='/hatchets/{{$name}}/charts/ops?type=stats&op={{$val.Name}}'; return false;"><i class='fa fa-area-chart'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "ip"}}
	<table style='float: left; margin: 10px 10px;'>
		<caption><button class='btn'
			onClick="javascript:location.href='/hatchets/{{.Hatchet}}/charts/connections?type=accepted'; return false;">
			<i class='fa fa-pie-chart'></i></button>Stats by IPs</caption>
		<tr><th></th><th>IP</th><th>Accepted Connections</th><th>Response Length</th></tr>
	{{range $n, $val := index .Data "ip"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:location.href='/hatchets/{{$name}}/charts/reslen-ip?ip={{$val.Name}}'; return false;"><i class='fa fa-pie-chart'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td><td align=right>{{getFormattedSize $val.Values 1}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "ns"}}
	<table style='float: left; margin: 10px 10px; clear: left;'>
		<caption><button class='btn'
			onClick="javascript:location.href='/hatchets/{{.Hatchet}}/charts/reslen-ns?ns='; return false;">
			<i class='fa fa-pie-chart'></i></button>Stats by Namespaces</caption>
		<tr><th></th><th>Namespace</th><th>Accessed</th><th>Response Length</th></tr>
	{{range $n, $val := index .Data "ns"}}
		<tr><td align=right>{{add $n 1}}</td>
		<td>
			<button class='btn' onClick="javascript:location.href='/hatchets/{{$name}}/logs/all?context={{$val.Name}}'; return false;"><i class='fa fa-search'></i></button>{{$val.Name}}
		</td>
		<td align=right>{{getFormattedNumber $val.Values 0}}</td><td align=right>{{getFormattedSize $val.Values 1}}</td></tr>
	{{end}}
	</table>
{{end}}

{{if hasData .Data "duration"}}
	<table style='float: left; margin: 10px 10px; clear: left;'>
		<caption><span style="font-size: 16px; padding: 5px 5px;"><i class="fa fa-shield"></i></span>Top N Long Lasting Connections</caption>
		<tr><th></th><th>Context</th><th>Duration</th></tr>
	{{range $n, $val := index .Data "duration"}}
			<tr><td align=right>{{add $n 1}}</td>
				<td><button class='btn' onClick="javascript:location.href='/hatchets/{{$name}}/logs/all?context={{getContext $val.Name}}'; return false;">
					<i class='fa fa-search'></i></button>{{$val.Name}}
				</td>
				<td align=right>{{getFormattedDuration $val.Values 0}}</td>
			</tr>
	{{end}}
	</table>
{{end}}
	<div style='clear: left;' align='center'><hr/><p/>@simagix</div>
	`
	html += "</body></html>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"hasData": func(data map[string][]NameValues, key string) bool {
			return len(data[key]) > 0
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
		"getFormattedNumber": func(numbers []int, i int) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", numbers[i])
		},
		"coinToss": func() bool {
			rand.Seed(time.Now().UnixNano())
			randomNum := rand.Intn(2)
			return (randomNum%2 == 0)
		},
		"getDurationFromSeconds": func(s int) string {
			return gox.GetDurationFromSeconds(float64(s))
		},
		"getFormattedDuration": func(numbers []int, i int) string {
			return gox.GetDurationFromSeconds(float64(numbers[i]))
		},
		"getStorageSize": func(s int) string {
			return gox.GetStorageSize(float64(s))
		},
		"getFormattedSize": func(numbers []int, i int) string {
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
				if info.Module != "enterprise" {
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
					html += "You should confirm the driver you use is <mark>compatible with the MongoDB server version</mark>. "
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
					html += "You should double check the drivers you use are <mark>compatible with the MongoDB server version</mark>. "
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
						conns += doc.Values[0]
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
						count += doc.Values[0]
					}
					html += printer.Sprintf("As many as <span style='color: orange;'>%d</span> different namespaces were accessed a total of <span style='color: orange;'>%d</span> times. ", len(docs), count)
					reslen := 0
					for _, doc := range docs {
						reslen += doc.Values[1]
					}
					html += printer.Sprintf("The total response length was around <span style='color: orange;'>%v</span>. ", gox.GetStorageSize(reslen))
				} else if key == "stats" && len(docs) > 0 {
					for _, doc := range docs {
						if doc.Name == "maxConns" {
							milli := doc.Values[0]
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
						} else if doc.Name == "maxMilli" && doc.Values[0] > 0 {
							milli := doc.Values[0]
							html += "The slowest operation took "
							if milli < 1000 {
								html += printer.Sprintf("<span style='color: orange;'>%d</span> milliseconds. ", doc.Values[0])
							} else {
								seconds := float64(milli) / 1000
								html += printer.Sprintf("<span style='color: orange;'>%s</span>. ", gox.GetDurationFromSeconds(seconds))
							}
						} else if doc.Name == "avgMilli" && doc.Values[0] > 0 {
							milli := doc.Values[0]
							html += printer.Sprintf("Moreover, the average operation time was <span style='color: orange;'>%d</span> milliseconds", milli)
							if milli > 100 {
								html += `, where operation time <mark>greater than 100 milliseconds is, IMO, "slow"</mark>. `
							} else {
								html += ". "
							}
						} else if doc.Name == "totalMilli" && doc.Values[0] > 0 {
							seconds := float64(doc.Values[0]) / 1000
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
							seconds := float64(doc.Values[0]) / 1000
							html += printer.Sprintf(`The <i>COLLSCAN</i> caused a total of <span style='color: orange;'>%s</span> wasted. `, gox.GetDurationFromSeconds(seconds))
						}
					}
				}
			}
			if totalImpact > 0 {
				html += "OMG, the total impact from slowest operations was "
				html += printer.Sprintf("<span style='color: orange;'>%s</span>, this may be a problem of lacking resources.  Please review the sizing training videos below. ", gox.GetDurationFromSeconds(totalImpact))
				html += "<div style='padding: 10px'>"
				html += `<iframe width="336" height="189" src="https://www.youtube.com/embed/0AAMw_q1E4o" title="YouTube video player" style="margin-right: 5px;” frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>`
				html += `<iframe width="336" height="189" src="https://www.youtube.com/embed/n1wORkr_1xI" title="YouTube video player" style="margin-right: 5px;” frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>`
				html += "</div>"
			}
			if len(data) > 0 {
				html += "<p/>Check out the details below for additional information about this server."
			}
			return template.HTML(html)
		}}).Parse(html)
}
