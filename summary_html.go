// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"html/template"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetSummaryTemplate returns HTML
func GetSummaryTemplate() (*template.Template, error) {
	return template.New("hatchet").Funcs(template.FuncMap{
		"numPrinter": func(n interface{}) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", ToInt(n))
		}}).Parse(html)
}

const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Ken Chen's Hatchet</title>
	<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
	<meta http-equiv="Pragma" content="no-cache" />
	<meta http-equiv="Expires" content="0" />

  <script src="https://www.gstatic.com/charts/loader.js"></script>
  <link href="/favicon.ico" rel="icon" type="image/x-icon" />
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">
  <style>
  	body {
		font-family: Helvetica, Arial, sans-serif;
		background-color: #f2f2f2;
		margin-top: 5px;
		margin-bottom: 10px;
		margin-right: 20px;
		margin-left: 20px;
  	}
    table
    {
    	font-family: Consolas, monaco, monospace;
    	border-collapse:collapse;
    	min-width:600px;
    }
    caption
    {
    	caption-side:top;
    	font-weight:bold;
    	font-style:italic;
    	margin:2px;
    }
    table, th, td
    {
		border: 1px solid gray;
		vertical-align: top;
    }
    th, td
    {
    	padding:2px;
    	vertical-align: top;
    }
    th
    {
      background-color: #ddd;
      font-weight:bold;
    }
    tr:nth-child(even) {background-color: #f2f2f2;}
    tr:nth-child(odd) {background-color: #fff;}
    .rowtitle
    {
    	font-weight:bold;
    }
	a {
	  text-decoration: none;
	  color: #000;
	  display: block;

	  -webkit-transition: font-size 0.3s ease, background-color 0.3s ease;
	  -moz-transition: font-size 0.3s ease, background-color 0.3s ease;
	  -o-transition: font-size 0.3s ease, background-color 0.3s ease;
	  -ms-transition: font-size 0.3s ease, background-color 0.3s ease;
	  transition: font-size 0.3s ease, background-color 0.3s ease;
	}
		a:hover {
			color: blue;
		}
    .button {
	  font-family: "Trebuchet MS";
      background-color: #4285f4;
      border: 2px solid #4285f4;
      border-radius: 4px;
      color: white;
      padding: 2px 4px;
      text-align: center;
      text-decoration: none;
      display: inline-block;
      font-size: 14px;
    }
    .fixed {
      position: fixed;
      top: 20px;
      right: 20px;
    }

    h1 {
	  font-family: "Trebuchet MS";
      font-size: 1.7em;
      font-weight: bold;
    }
    h2 {
	  font-family: "Trebuchet MS";
      font-size: 1.5em;
      font-weight: bold;
    }
    h3 {
	  font-family: "Trebuchet MS";
      font-size: 1.25em;
      font-weight: bold;
    }
    h4 {
	  font-family: "Trebuchet MS";
      font-size: 1em;
      font-weight: bold;
	}
	.command {
	  background-color: #fff;
	  border: none;
	  outline:none;
	}
	.btn {
	  background-color: #fff;
	  border: none;
	  outline:none;
	  color: #4285F4;
	  padding: 5px 30px;
	  cursor: pointer;
	  font-size: 20px;
	}
	.btn:hover {
	  color: blue;
	  border: none;
	}
    </style>
</head>

<body>
<script type="text/javascript">
	function toggleDiv(tag) {
		var x = document.getElementById(tag);
		if (x.style.display === "none") {
	  		x.style.display = "block";
		} else {
	  		x.style.display = "none";
		}
  	}
</script>

<div align='center'>
	<table width='100%'>
		<tr>
			<th>command</th>
			<th>namespace</th>
			<th>count</th>
			<th>avg ms</th>
			<th>max ms</th>
			<th>total ms</th>
			<th>total reslen</th>
			<th>index</th>
			<th>query pattern</th>
		</tr>
{{range $key, $value := .Ops}}
		<tr>
			<td>{{ $value.Command }}</td>
			<td>{{ $value.Namespace }}</td>
			<td align='right'>{{ numPrinter $value.Count }}</td>
			<td align='right'>{{ numPrinter $value.AvgMilli }}</td>
			<td align='right'>{{ numPrinter $value.MaxMilli }}</td>
			<td align='right'>{{ numPrinter $value.TotalMilli }}</td>
			<td align='right'>{{ numPrinter $value.Reslen }}</td>
		{{ if ( eq $value.Plan "COLLSCAN" ) }}
			<td><span style="color:red;">{{ $value.Plan }}</span></td>
		{{else}}
			<td>{{ $value.Index }}</td>
		{{end}}
			<td>{{ $value.Filter }}</td>
		</tr>
{{end}}
	</table>
</div>
</body>
`
