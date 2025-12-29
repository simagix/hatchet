/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * template.go
 */

package hatchet

import (
	"fmt"
	"html/template"
	"sort"
)

// GetTablesTemplate returns HTML
func GetTablesTemplate() (*template.Template, error) {
	html := headers + getMainPage() + "</body>"
	return template.New("hatchet").Funcs(template.FuncMap{
		"getHatchetImage": func() string {
			return HATCHET_PNG
		},
		"add": func(a int, b int) int {
			return a + b
		}}).Parse(html)
}

const headers = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Ken Chen's Hatchet</title>
	<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
	<meta http-equiv="Pragma" content="no-cache" />
	<meta http-equiv="Expires" content="0" />

  <script src="https://www.gstatic.com/charts/loader.js"></script>
  <link href="/favicon.ico" rel="icon" type="image/x-icon" />
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Righteous&display=swap" rel="stylesheet">
  <style>
    :root {
      --text-color: #2C5234;
      --header-color: #D5E0D5;
      --row-color: #F5F8F5;
      --background-color: #FFF9F0;
      --accent-color-1: #7BAF9B;
      --accent-color-2: #9FCCB3;
      --accent-color-3: #3D6B47;
      --border-color: #C5D5C5;
    }
  	body {
      font-family: Helvetica, Arial, sans-serif;
      margin-top: 10px;
      margin-bottom: 10px;
      margin-right: 10px;
      margin-left: 10px;
	    background-color: var(--background-color);
    }
    table {
      border-collapse:collapse;
      min-width: 300px;
    }
    caption {
      caption-side: top;
      font-size: 1.1em;
	    text-align: left;
    }
    table, th, td {
      border: 1px solid var(--border-color);
      vertical-align: middle;
    }
    th {
      background-color: var(--header-color);
      color: var(--text-color);
      font-weight: bold;
      padding: 8px 10px;
      font-size: 1.1em;
      text-align: left;    
    }
    td {
      background-color: var(--row-color);
      padding: 8px 10px;
      font-size: 1.1em;
    }
    tr:nth-child(even) td {
      background-color: white;
    }
    .break {
      vertical-align: middle;
      font-size: 1.05em;
      word-break: break-all;
    }
    table a:link {
      color: var(--text-color);
      text-decoration: none;
    }
    table a:visited {
      color: var(-accent-color-1);
      text-decoration: none;
    }
    table a:hover {
      color: red;
      text-decoration: none;
    }
    ul, ol {
      #font-family: Consolas, monaco, monospace;
      font-size: 1em;
    }
    .btn {
      background-color: transparent;
      border: none;
      outline:none;
      color: var(--accent-color-3);
      padding: 2px 2px;
      cursor: pointer;
      font-size: 18px;
      font-weight: bold;
      border-radius: .25em;
    }
    .btn:hover {
      background-color: var(--border-color);
      color: #DB4437;
    }
    .download-btn {
      background-color: #e8efe8;
      border: 1px solid #ccc;
      color: #666;
      padding: 6px 12px;
      border-radius: 4px;
      cursor: pointer;
      font-size: 16px;
      transition: all 0.2s ease;
    }
    .download-btn:hover {
      background-color: var(--accent-color-3);
      color: white;
      border-color: var(--accent-color-3);
    }
    .button { 
      background-color: var(--text-color);
      border: 1px solid #999; 
      outline: none;
      color: var(--background-color);
      padding: 8px 15px;
      margin: 0px 10px;
      cursor: pointer;
      font-size: 16px;
      font-weight: bold;
      border-radius: 4px;
      box-sizing: border-box;
    }
    .exclamation {
      background: none;
      color: red;
      border: none;
      outline: none;
      padding: 5px 10px;
      margin: 2px 2px;
      font-size: 1em;
      border-radius: .25em;
    }
    .tooltip {
      position: relative;
      display: inline-block;
    }
    .tooltip .tooltiptext {
      visibility: hidden;
      width: 200px;
      background-color: #555;
      color: #fff;
      text-align: left;
      border-radius: 6px;
      padding: 5px 5px;
      position: absolute;
      z-index: 1;
      bottom: 125%;
      left: 50%;
      margin-left: -100px;
      opacity: 0;
      transition: opacity 0.3s;
    }
    .tooltip .tooltiptext::after {
      content: "";
      position: absolute;
      top: 100%;
      left: 50%;
      margin-left: -5px;
      border-width: 5px;
      border-style: solid;
      border-color: #555 transparent transparent transparent;
    }
    .tooltip:hover .tooltiptext {
      visibility: visible;
      opacity: 1;
    }
    h1 {
      font-size: 1.8em;
      font-weight: bold;
    }
    h2 {
      font-size: 1.6em;
      font-weight: bold;
    }
    h3 {
      font-size: 1.3em;
      font-weight: bold;
    }
    h4 {
      font-size: 1.1em;
      font-weight: bold;
    }
    .footer {
      background-color: #fff;
      opacity: .75;
      position: fixed;
      left: 0;
      bottom: 0;
      width: 100%;
      color: #000;
      text-align: left;
      padding: 2px 10px;
    }
    input, select, textarea {
      font-family: "Trebuchet MS";
      appearance: auto;
      background-color: #f2f2f2;
      color: var(--text-color);
      border-radius: 4px;
      font-size: 16px;
      padding: 8px 12px;
      border: 1px solid #999;
    }
    select {
      font-weight: bold;
      font-size: 17px;
    }
    .rotate23:hover {
      -webkit-transform: rotate(23deg);
      -moz-transform: rotate(23deg);
      -o-transform: rotate(23deg);
      -ms-transform: rotate(23deg);
      transform: rotate(23deg);
    }
    input[type="checkbox"] {
      accent-color: red;
    }
    .sort {
      color: #FFF;
    }
    .sort:hover {
      color: #DB4437;
    }
    .summary {
      font-family: Consolas, monaco, monospace;
	    background-color: #111;
      color: var(--border-color);
	    padding: .5rem;
	    margin: .5rem;
      font-size: 1em;
    }
    #loading {
      position: fixed;
      top: 0;
      left: 0;
      bottom: 0;
      right: 0;
      background-color: rgba(0, 0, 0, 0.5);
      z-index: 9999;
	  display: none;
    }
    .spinner {
      border: 5px solid #f3f3f3;
      border-top: 5px solid #3498db;
      border-radius: 50%;
      width: 50px;
      height: 50px;
      animation: spin 2s linear infinite;
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
    }
    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }
    .chart {
      background-color: var(--row-color);;
      border: solid;
      padding: 10px 10px;
      margin: 10px 10px;
      border-radius: .5em;
    }
  </style>
  <script>
    function loadData(url, skipHistory) {
    	var loading = document.getElementById('loading');
    	loading.style.display = 'block';
    	fetch(url, {
    		headers: { 'X-Requested-With': 'XMLHttpRequest' }
    	})
        	.then(response => response.text())
        	.then(data => {
      			loading.style.display = 'none';
      			if (!skipHistory) {
      				history.pushState({url: url}, '', url);
      			}
				document.open();
				document.write(data);
				document.close();
        	})
        	.catch(error => {
      			loading.style.display = 'none';
        	});
    }
    // Handle browser back/forward buttons
    window.onpopstate = function(event) {
    	if (event.state && event.state.url) {
    		loadData(event.state.url, true);
    	} else {
    		location.href = '/';
    	}
    };
  </script>
</head>
<body>
  <div id="loading">
    <div class="spinner"></div>
  </div>
`

func getContentHTML() string {
	html := headers
	html += `
<style>
  .menu-bar { display: flex; align-items: center; padding: 8px 0; border-bottom: 1px solid var(--border-color); margin-bottom: 10px; }
  .menu-item { display: inline-flex; align-items: center; gap: 8px; padding: 10px 16px; margin-right: 8px; font-size: 17px; font-weight: bold; color: var(--text-color); background: var(--header-color); border: none; border-radius: 4px; cursor: pointer; text-decoration: none; }
  .menu-item:hover { background: var(--accent-color-3); color: white; }
  .menu-item.active { background: var(--text-color); color: white; }
  .menu-item i { font-size: 18px; }
  .menu-right { margin-left: auto; display: flex; align-items: center; gap: 8px; }
  .menu-dropdown { position: relative; display: inline-block; }
  .menu-dropdown-content { display: none; position: absolute; right: 0; background: #fff; min-width: 220px; box-shadow: 0 4px 12px rgba(0,0,0,0.15); border-radius: 6px; z-index: 100; max-height: 400px; overflow-y: auto; }
  .menu-dropdown-content a { display: block; padding: 10px 16px; color: var(--text-color); text-decoration: none; font-size: 15px; }
  .menu-dropdown-content a:hover { background: var(--header-color); }
  .menu-dropdown:hover .menu-dropdown-content { display: block; }
  .menu-dropdown:hover .menu-item { background: var(--accent-color-3); color: white; }
  .content-container { background: #EEF2EE; border-radius: 24px; padding: 30px 40px; margin: 20px auto; box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08); }
</style>
<div class="menu-bar">
  <button class="menu-item" data-page="home" onclick="location.href='/'; return false;">
    <i class="fa fa-home"></i> Hatchet
  </button>
  <button class="menu-item" data-page="audit" onclick="loadData('/hatchets/{{.Hatchet}}/stats/audit'); return false;">
    <i class="fa fa-shield"></i> Audit
  </button>
  <button class="menu-item" data-page="stats" onclick="loadData('/hatchets/{{.Hatchet}}/stats/slowops'); return false;">
    <i class="fa fa-info"></i> Stats
  </button>
  <button class="menu-item" data-page="topn" onclick="loadData('/hatchets/{{.Hatchet}}/logs/slowops'); return false;">
    <i class="fa fa-list"></i> Top N
  </button>
  <div class="menu-dropdown">
    <button class="menu-item" data-page="charts">
      <i class="fa fa-bar-chart"></i> Charts <i class="fa fa-caret-down" style="margin-left: 4px;"></i>
    </button>
    <div class="menu-dropdown-content">`
	items := []Chart{}
	for _, chart := range charts {
		items = append(items, chart)
	}
	sort.Slice(items, func(i int, j int) bool {
		return items[i].Index < items[j].Index
	})

	for i, item := range items {
		if i == 0 {
			continue
		}
		html += fmt.Sprintf("<a href='#' onclick=\"loadData('/hatchets/{{.Hatchet}}/charts%v'); return false;\">%v</a>", item.URL, item.Title)
	}

	html += `</div>
  </div>
  <button class="menu-item" data-page="search" onclick="loadData('/hatchets/{{.Hatchet}}/logs/all?component=NONE'); return false;">
    <i class="fa fa-search"></i> Search
  </button>
</div>
<script>
	function setChartType() {
		// No longer needed - charts dropdown is now in menu
	}

	function refreshChart() {
		var sd = document.getElementById('start').value;
		var ed = document.getElementById('end').value;
		loadData('/hatchets/{{.Hatchet}}/charts{{.Chart.URL}}&duration=' + sd + ',' + ed);
	}

	// Highlight active menu item based on URL
	(function() {
		var path = window.location.pathname;
		var page = 'home';
		if (path.includes('/stats/audit')) page = 'audit';
		else if (path.includes('/stats/slowops')) page = 'stats';
		else if (path.includes('/logs/slowops')) page = 'topn';
		else if (path.includes('/logs/all')) page = 'search';
		else if (path.includes('/charts/')) page = 'charts';
		var btn = document.querySelector('.menu-item[data-page="' + page + '"]');
		if (btn) btn.classList.add('active');
	})();
</script>
<div class="content-container">
`
	html += getFooter()
	return html
}

func getMainPage() string {
	template := `
<script>
	function selectHatchet(name) {
		loadData('/hatchets/' + name + '/stats/audit'); 
	}
	function renameHatchet(oldName, event) {
		event.stopPropagation();
		var newName = prompt("Enter new name:", oldName);
		if (newName && newName !== oldName) {
			var loading = document.getElementById('loading');
			loading.style.display = 'block';
			fetch('/api/hatchet/v1.0/rename?old=' + encodeURIComponent(oldName) + '&new=' + encodeURIComponent(newName), {method: 'POST'})
				.then(response => response.json())
				.then(data => {
					loading.style.display = 'none';
					if (data.ok) {
						loadData('/');
					} else {
						alert('Error: ' + data.error);
					}
				})
				.catch(error => {
					loading.style.display = 'none';
					alert('Error: ' + error);
				});
		}
	}
	function deleteHatchet(name, event) {
		event.stopPropagation();
		if (confirm('Delete "' + name + '"?\n\nThis action cannot be undone.')) {
			var loading = document.getElementById('loading');
			loading.style.display = 'block';
			fetch('/api/hatchet/v1.0/delete?name=' + encodeURIComponent(name), {method: 'DELETE'})
				.then(response => response.json())
				.then(data => {
					loading.style.display = 'none';
					if (data.ok) {
						loadData('/');
					} else {
						alert('Error: ' + data.error);
					}
				})
				.catch(error => {
					loading.style.display = 'none';
					alert('Error: ' + error);
				});
		}
	}

	// Upload handling
	function setupUploadZone() {
		var zone = document.getElementById('upload-zone');
		if (!zone) return;
		
		zone.addEventListener('dragover', function(e) {
			e.preventDefault();
			zone.classList.add('dragover');
		});
		zone.addEventListener('dragleave', function(e) {
			e.preventDefault();
			zone.classList.remove('dragover');
		});
		zone.addEventListener('drop', function(e) {
			e.preventDefault();
			zone.classList.remove('dragover');
			var files = e.dataTransfer.files;
			if (files.length > 0) {
				handleFileSelect(files[0]);
			}
		});
	}
	
	function handleFileSelect(file) {
		if (!file) return;
		
		var zone = document.getElementById('upload-zone');
		var status = document.getElementById('upload-status');
		
		zone.classList.add('uploading');
		status.innerHTML = '<div class="progress-container"><div class="progress-bar" id="upload-progress"></div></div><div id="progress-text">Uploading ' + file.name + '... 0%</div>';
		
		var formData = new FormData();
		formData.append('logfile', file);
		
		var xhr = new XMLHttpRequest();
		xhr.open('POST', '/api/hatchet/v1.0/upload', true);
		
		xhr.upload.onprogress = function(e) {
			if (e.lengthComputable) {
				var percent = Math.round((e.loaded / e.total) * 100);
				document.getElementById('upload-progress').style.width = percent + '%';
				document.getElementById('progress-text').textContent = 'Uploading ' + file.name + '... ' + percent + '%';
			}
		};
		
		xhr.onload = function() {
			zone.classList.remove('uploading');
			try {
				var data = JSON.parse(xhr.responseText);
				if (data.status === 'processing') {
					status.innerHTML = '<i class="fa fa-cog fa-spin"></i> Processing ' + data.name + '...';
					pollUploadStatus(data.name);
				} else {
					status.innerHTML = '<i class="fa fa-times" style="color: red;"></i> Error: ' + (data.error || 'Upload failed');
				}
			} catch (e) {
				status.innerHTML = '<i class="fa fa-times" style="color: red;"></i> Error: ' + xhr.responseText;
			}
		};
		
		xhr.onerror = function() {
			zone.classList.remove('uploading');
			status.innerHTML = '<i class="fa fa-times" style="color: red;"></i> Upload failed';
		};
		
		xhr.send(formData);
	}
	
	function pollUploadStatus(name) {
		var status = document.getElementById('upload-status');
		var pollCount = 0;
		var maxPolls = 300; // 5 minutes max
		
		var poll = setInterval(function() {
			pollCount++;
			fetch('/api/hatchet/v1.0/upload/status/' + encodeURIComponent(name))
				.then(response => response.json())
				.then(data => {
					if (data.status === 'complete') {
						clearInterval(poll);
						status.innerHTML = '<i class="fa fa-check" style="color: green;"></i> Processing complete!';
						loadData('/'); // Refresh the page to show new hatchet
					} else if (pollCount >= maxPolls) {
						clearInterval(poll);
						status.innerHTML = '<i class="fa fa-clock-o"></i> Still processing... refresh page to check.';
					} else {
						status.innerHTML = '<i class="fa fa-spinner fa-spin"></i> Processing... (' + pollCount + 's)';
					}
				})
				.catch(error => {
					// Keep polling on error
				});
		}, 1000);
	}
	
	// Initialize upload zone on page load
	document.addEventListener('DOMContentLoaded', setupUploadZone);
	// Convert UTC time to browser local time
	function convertToLocalTime() {
		document.querySelectorAll('.utc-time').forEach(function(el) {
			var utc = el.getAttribute('data-utc');
			if (utc && utc.length > 0) {
				// Parse UTC time (format: YYYY-MM-DD HH:MM:SS)
				var d = new Date(utc.replace(' ', 'T') + 'Z');
				if (!isNaN(d.getTime())) {
					el.textContent = d.toLocaleString(undefined, { hour12: false });
				}
			}
		});
	}
	window.onload = convertToLocalTime;
</script>
<style>
	.hatchet-table-container {
		max-height: 200px;
		overflow-y: auto;
		border: 1px solid var(--border-color);
		border-radius: 4px;
	}
	.hatchet-table {
		width: 100%;
		margin: 0;
		border: none;
	}
	.hatchet-table th {
		position: sticky;
		top: 0;
		z-index: 1;
	}
	.hatchet-table tr.clickable-row {
		cursor: pointer;
	}
	.hatchet-table tr.clickable-row:hover td {
		background-color: var(--accent-color-3) !important;
		color: white;
	}
	.hatchet-table tr.clickable-row:hover .action-btn {
		color: #FFD700;
	}
	.hatchet-table tr.clickable-row:hover .delete-btn {
		color: #FF6B6B;
	}
	.upload-zone {
		border: 2px dashed #999;
		border-radius: 8px;
		padding: 20px;
		text-align: center;
		cursor: pointer;
		background: #f9f9f9;
		transition: all 0.2s ease;
		min-height: 120px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
	}
	.upload-zone:hover, .upload-zone.dragover {
		border-color: var(--accent-color-3);
		background: #f0f5f0;
	}
	.upload-zone.uploading {
		opacity: 0.7;
		pointer-events: none;
	}
	.progress-container {
		width: 100%;
		height: 8px;
		background: #ddd;
		border-radius: 4px;
		overflow: hidden;
		margin-bottom: 6px;
	}
	.progress-bar {
		height: 100%;
		width: 0%;
		background: var(--accent-color-3);
		transition: width 0.2s ease;
	}
	#progress-text {
		font-size: 0.9em;
		color: #666;
	}
	.action-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 2px 6px;
		color: #E8B923;
		font-size: 14px;
		border-radius: 3px;
	}
	.action-btn:hover {
		color: #D4A017;
	}
	.delete-btn {
		color: #CC5555;
	}
	.delete-btn:hover {
		color: #FF4444;
	}
	.help-link {
		color: var(--accent-color-3);
		cursor: pointer;
		font-size: 1em;
		text-decoration: none;
	}
	.help-link:hover {
		color: var(--text-color);
		text-decoration: underline;
	}
	.help-link i {
		margin-right: 6px;
	}
	.help-section {
		display: none;
		margin-top: 15px;
		border: 1px solid var(--border-color);
		border-radius: 6px;
		background: #fff;
		padding: 16px;
	}
	.help-section.open {
		display: block;
	}
	.help-section h3:first-child {
		margin-top: 0;
	}
	.home-container {
		background: #EEF2EE;
		border-radius: 24px;
		padding: 30px 40px;
		margin: 30px auto;
		max-width: 1200px;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
	}
</style>

<div class="home-container">
<!-- Row 1: Title/Description (50%) | Tutorial Video (50%) -->
<div style='display: flex; gap: 30px; padding: 15px 0; border-bottom: 1px solid var(--border-color); margin-bottom: 15px;'>
	<div style='flex: 1; display: flex; flex-direction: column; justify-content: flex-start;'>
		<h1 style='margin: 0 0 16px 0; font-size: 3.2em; font-family: Righteous, cursive; letter-spacing: 3px; color: #333;'>Hatchet<img src='data:image/png;base64,` + CHEN_ICO + `' style='vertical-align: top; margin-left: 2px; transform: rotate(-23deg); position: relative; top: -5px;'/></h1>
		<p style='margin: 0; color: #444; font-size: 1.35em; line-height: 1.6; max-width: 560px; font-style: italic;'>
			Like a skilled woodsman reading the rings of a tree, Hatchet reveals the stories hidden within your MongoDB logs — from performance patterns and activity rhythms to security insights and troubleshooting trails.
		</p>
	</div>
	<div style='flex: 1; display: flex; align-items: center; justify-content: center;'>
		<iframe width="560" height="315" src="https://www.youtube.com/embed/WavOyaFTDE8" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>
	</div>
</div>

<!-- Row 2: Upload and Hatcheted Logs Selection (full width) -->
<div style='margin-bottom: 15px;'>
	<div style='display: flex; gap: 20px; align-items: flex-start;'>
		<!-- Upload Section -->
		<div style='flex: 0 0 280px;'>
			<label style='font-weight: bold; font-size: 1.3em; margin-bottom: 8px; display: block;'>Upload log file:</label>
			<div id='upload-zone' class='upload-zone' onclick='document.getElementById("file-input").click()'>
				<i class='fa fa-cloud-upload' style='font-size: 2em; color: #666; margin-bottom: 8px;'></i>
				<div>Drop file here or click to browse</div>
				<div style='font-size: 0.85em; color: #888; margin-top: 4px;'>MongoDB log files, .gz supported (max 200 MB)</div>
				<input type='file' id='file-input' style='display: none;' onchange='handleFileSelect(this.files[0])'>
			</div>
			<div id='upload-status' style='margin-top: 8px; font-size: 0.9em;'></div>
		</div>
		<!-- Hatcheted Logs Table -->
		<div style='flex: 1;'>
			<label style='font-weight: bold; font-size: 1.3em; margin-bottom: 8px; display: block;'>Select a hatcheted log:</label>
			<div class='hatchet-table-container'>
				<table class='hatchet-table'>
					<tr><th>#</th><th>Hatcheted Log</th><th>Processed Time</th></tr>
{{range $n, $entry := .Hatchets}}
					<tr class='clickable-row' onclick='selectHatchet("{{$entry.Name}}")'>
						<td style='text-align: center; width: 40px;'>{{add $n 1}}</td>
						<td><button class='action-btn' onclick='renameHatchet("{{$entry.Name}}", event)' title='Rename'><i class='fa fa-pencil'></i></button><button class='action-btn delete-btn' onclick='deleteHatchet("{{$entry.Name}}", event)' title='Delete'><i class='fa fa-trash'></i></button> {{$entry.Name}}</td>
						<td class='utc-time' data-utc='{{$entry.CreatedAt}}'>{{$entry.CreatedAt}}</td>
					</tr>
{{else}}
					<tr><td colspan='3' style='text-align: center; color: #666;'>No logs processed yet</td></tr>
{{end}}
				</table>
			</div>
		</div>
	</div>
</div>

<a class="help-link" onclick="document.getElementById('help-section').classList.toggle('open'); this.querySelector('span').textContent = document.getElementById('help-section').classList.contains('open') ? 'Hide documentation' : 'View documentation'; return false;">
	<i class="fa fa-question-circle"></i> <span>View documentation</span> →
</a>
<div id="help-section" class="help-section">
<h3>Reports</h3>
    <table width='100%'>
      <tr><th></th><th>Title</th><th>Description</th></tr>
      <tr><td align=center><i class="fa fa-shield"></i></td><td>Audit</td><td>Display information on security audits and performance metrics</td></tr>
      <tr><td align=center><i class="fa fa-bar-chart"></i></td><td>Charts</td><td>A number of charts are available for security audits and performance metrics</td></tr>
      <tr><td align=center><i class="fa fa-search"></i></td><td>Search</td><td>Powerful log searching function with key metrics highlighted</td></tr>
      <tr><td align=center><i class="fa fa-info"></i></td><td>Stats</td><td>Summary of slow operational query patterns and duration</td></tr>
      <tr><td align=center><i class="fa fa-list"></i></td><td>TopN</td><td>Display the slowest 23 operation logs</td></tr>
    </table>
<h3 style='margin-top: 24px;'>Charts</h3>
    <table width='100%'>
      <tr><th></th><th>Title</th><th>Description</th></tr>`
	size := len(charts) - 1
	tables := make([]Chart, size)
	for k, chart := range charts {
		if k == "instruction" {
			continue
		}
		tables[chart.Index-1] = chart
	}
	for _, chart := range tables {
		template += fmt.Sprintf("<tr><td align=right>%d</td><td>%v</td><td>%v</td></tr>\n",
			chart.Index, chart.Title, chart.Descr)
	}
	template += "</table>"
	template += `<h3 style='margin-top: 24px;'>URL</h3>
<ul class="api">
	<li>/</li>
	<li>/hatchets/{hatchet}/charts/{chart}[?type={str}]</li>
	<li>/hatchets/{hatchet}/logs/all[?component={str}&context={str}&duration={date},{date}&severity={str}&limit=[{offset},]{int}]</li>
	<li>/hatchets/{hatchet}/logs/slowops[?topN={int}]</li>
	<li>/hatchets/{hatchet}/stats/slowops[?COLLSCAN={bool}&orderBy={str}]</li>
</ul>

<h3 style='margin-top: 24px;'>API</h3>
<ul class="api">
	<li><b>POST</b> /api/hatchet/v1.0/upload - Upload log file (multipart form: logfile, name)</li>
	<li><b>GET</b> /api/hatchet/v1.0/upload/status/{name} - Check upload processing status</li>
	<li><b>POST</b> /api/hatchet/v1.0/rename?old={name}&new={name} - Rename a hatchet</li>
	<li><b>DELETE</b> /api/hatchet/v1.0/delete?name={name} - Delete a hatchet</li>
	<li>/api/hatchet/v1.0/hatchets/{hatchet}/logs/all[?component={str}&context={str}&duration={date},{date}&severity={str}&limit=[{offset},]{int}]</li>
	<li>/api/hatchet/v1.0/hatchets/{hatchet}/logs/slowops[?topN={int}]</li>
	<li>/api/hatchet/v1.0/hatchets/{hatchet}/stats/audit</li>
	<li>/api/hatchet/v1.0/hatchets/{hatchet}/stats/slowops[?COLLSCAN={bool}&orderBy={str}]</li>
	<li>/api/hatchet/v1.0/mongodb/{version}/drivers/{driver}?compatibleWith={driver version}</li>
</ul>
</div>
</div><!-- end home-container -->
<h4 align='center'><hr/>{{.Version}}</h4>
`
	template += fmt.Sprintf(`<div class="footer"><img valign="middle" src='data:image/png;base64,%v'/> Ken Chen</div>`, CHEN_ICO)
	return template
}

// GetErrorTemplate returns an error page template
func GetErrorTemplate() (*template.Template, error) {
	html := headers + `
<div style='max-width: 600px; margin: 50px auto; text-align: center;'>
	<h1 style='color: #DB4437; font-size: 2em;'><i class="fa fa-exclamation-triangle"></i> Error</h1>
	<div style='background: #FFF; border: 1px solid var(--border-color); border-radius: 8px; padding: 20px; margin: 20px 0;'>
		<p style='color: #666; font-size: 1.1em; margin-bottom: 15px;'>{{.Message}}</p>
		<p style='color: #999; font-size: 0.9em;'>Hatchet: <strong>{{.Hatchet}}</strong></p>
	</div>
	<button class='button' onclick="location.href='/';" style='font-size: 1.1em; padding: 10px 30px;'>
		<i class="fa fa-home"></i> Back to Home
	</button>
</div>
</body></html>`
	return template.New("error").Parse(html)
}
