# Hatchet Developer's Guide
The Hatchet provides many flexible ways to access data.  After processing logs, Hatchet stores processed data in the embedded SQLite3 databae and provide a web interface to render reports.  In summary, you can use the methods below to view results or retrieve data to be used in your application.  The available methods are:

- View available HTML reports using a browser
- Use Hatchet RESTful APIs to import JSON data into applications to create reports
- Query the SQLite3 database using the `sqlite3` shell
- Access data directly in applications using SQLite3 API
- Output data from the SQLite3 database to a TSV file to be used in a spreadsheet
- Output the legacy-formatted logs to a file for other tools
- Can serve as a RESTful to check drivers compatibility
- Supported databases are SQLite3 and MongoDB (with `-url {URL}`)

Note that there are a few indexes created during the logs processings  But, you may create additional indexes to support additional needs.

## View Available Reports
The easiest way is to go to the home page `http://localhost:3721` and following the instructions to view available reports.  Each report is also available using its own URL with additional parameters defined in the query string.  Below are a few examples:

- `/hatchets/{hatchet}/stats/audit` view audit data
- `/hatchets/{hatchet}/stats/slowops?COLLSCAN=true&orderBy=count` views stats summary of COLLSCAN logs and sorted by *count*
- `/hatchets/{hatchet}/logs/slowops` views top 23 slowest ops logs
- `/hatchets/{hatchet}/logs/slowops?topN=100` views top 100 slowest ops logs
- `/hatchets/{hatchet}/logs/all` views all logs, and available query string parameters are:
  - component
  - context
  - duration (begin_datetime,end_datetime)
  - limit ([offset,]limit)
  - severity
- `/hatchets/{hatchet}/logs/all?component=NETWORK` searches logs where *component* = *NETWORK*.  Available option are:
  - component
  - context
  - duration (begin_datetime,end_datetime)
  - severity
- `/hatchets/{hatchet}/charts/connections[?type={}]` views connections charts, types are:
  - accepted
  - time
  - total
- `/hatchets/{hatchet}/charts/ops?type={}` views average ops time chart, types are:
  - stats
  - counts
- `/hatchets/{hatchet}/charts/reslen-ip?ip={}` views response length by IPs chart, types are:
- `/hatchets/{hatchet}/charts/reslen-ns?ns={}` views response length by IPs chart, types are:
```

## Query SQLite3 Database
The database file is *data/hatchet.db*; use the *sqlite3* command as below:
```bash
sqlite3 ./data/hatchet.db
```

After a log file is processed, 5 tables are created in the SQLite3 database.  The table name is derived from the parent directory and log file name.  For example, processing *rs1/mongod.log.gz* creates a table named *rs1_mongod*, while processing *rs2/mongod.log.gz* creates *rs2_mongod*.  This allows logs from replica set members with the same filename to be stored separately.  If a name collision still occurs, a sequential suffix (_2, _3, etc.) is added.  The other 4 tables are 1) {name}_ops stores stats of slow ops, 2) {name}_clients stores clients information, 3) {name}_audit keeps audit data, and 4) {name}_drivers to store driver information.  Re-processing the same log file will replace the existing data.  A few SQL commands follow.

### Query All Data
```sqlite3
SELECT * FROM mongod;
```

```sqlite3
SELECT date, severity, component, context, SUBSTR(message, 1, 60) message FROM mongod;
```

```sqlite3
SELECT date, severity, message FROM mongod WHERE component = 'NETWORK';
```

### Query Ops Stats
```sqlite3
SELECT op, COUNT(*) "count", ROUND(AVG(milli),1) avg_ms, MAX(milli) max_ms, SUM(milli) total_ms,
       ns, _index "index", SUM(reslen) "reslen", filter "query_pattern"
    FROM mongod
    WHERE op != "" GROUP BY op, ns, filter ORDER BY avg_ms DESC;
```

```sqlite3
SELECT SUBSTR(date, 1, 16), COUNT(op), op, ns, filter 
    FROM mongod where op != ''
    GROUP by SUBSTR(date, 1, 16), op, ns, filter;
```

```sqlite3
SELECT SUBSTR(date, 1, 16), COUNT(op), op, ns
    FROM mongod where op != ''
    GROUP by SUBSTR(date, 1, 16), op, ns;
```

## Use SQLite3 API
Different drivers are supported for most popular programming languages including Golang, NodeJS, Java, Python, and C#.

## Export TSV File
Export data to a TVS file and import it to a spreadsheet software.  Here is an example:
```bash
sqlite3 -header -separator $'\t' ./data/hatchet.db "SELECT * FROM mongod;" > mongod.tsv
```

## Hatchet API
Hatchet provides a number of APIs to output JSON data. They work similarly to the URLs but with a prefix `/api/hatchet/v1.0`.  The APIs are as follows:
- /api/hatchet/v1.0/hatchets/{hatchet}/stats/audit
- /api/hatchet/v1.0/hatchets/{hatchet}/stats/slowops[?orderyBy=] ; Possible values of *orderBy* are:
  - op
  - ns
  - count
  - avg_ms
  - max_ms
  - total_ms
  - reslen
- /api/hatchet/v1.0/hatchets/{hatchet}/logs/all
- /api/hatchet/v1.0/hatchets/{hatchet}/logs/slowops[?topN=] ; The default value of topN is 23.
- /api/hatchet/v1.0/mongodb/{version}/drivers/{driver}[?compatibleWith={driver version}]

## Output Logs in Legacy Format
```bash
./dist/hatchet -legacy testdata/mongod.log.gz > mongod_legacy.log
```

## In-Memory Mode
The in-memory mode is good for a quick view of the result and no data is persisted.  When using the in-memory mode, the web server is automatically started.  The in-memory mode is not necessarily faster than using a data file if the computer doesn't have enough memory.
```bash
./dist/hatchet -in-memory testdata/mongod.log.gz
```

## Docker Build
See Dockerfile for details.
