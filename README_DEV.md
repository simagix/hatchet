# Hatchet Developer's Guide
The Hatchet tool provides many different ways to access data and there are:
- Query the SQLite3 database
- Output to a TSV file and import data into a spreadsheet
- Use Hatchet API to import JSON data into applications
- View the provided HTML reports in a browser
- Output logs in the legacy format and processed by other tools
- Use SQLite3 API to access data from applications

Note that there are a few indexes created during the logs processings.  But,
you can create additional indexes to support additional needs.

## Query SQLite3 Database
The default database is stored in the *data/hatchet.db* file.
```bash
sqlite3 ./data/hatchet.db
```

### SQLite3 Command Examples
After a log file is processed, a table is created in the SQLite3 database.  The table name is 
part of the log file name.  A table name *mongod_json* is extracted from a log file name of, for example, 
$HOME/Downloads/**mongod_json**.log.gz.

### Query Data
```sqlite3
SELECT * from mongod_json;
```

```sqlite3
SELECT date, severity, component, context, substr(message, 1, 60) message FROM mongod_json;
```

```sqlite3
SELECT date, severity, message FROM mongod_json WHERE component = 'NETWORK';
```

### Query Performance Data
```sqlite3
SELECT op, COUNT(*) "count", ROUND(AVG(milli),1) avg_ms, MAX(milli) max_ms, SUM(milli) total_ms,
       ns, _index "index", SUM(reslen) "reslen", filter "query pattern"
    FROM mongod_json 
    WHERE op != "" GROUP BY op, ns, filter ORDER BY avg_ms DESC;
```

### Query Ops Activities
Use these commands to create charts.

#### Group by Op, Namespace, and Query Pattern
Query to form a profiler chart by grouping ops, namespaces, and query patterns in a 1-minute bucket.
```sqlite3
SELECT SUBSTR(date, 1, 16), COUNT(op), op, ns, filter 
    FROM mongod_json where op != ''
    GROUP by SUBSTR(date, 1, 16), op, ns, filter;
```

#### Group by Op and Namespace
Query to form a profiler chart by grouping ops and namespaces in a 1-minute bucket.
```sqlite3
SELECT SUBSTR(date, 1, 16), COUNT(op), op, ns
    FROM mongod_json where op != ''
    GROUP by SUBSTR(date, 1, 16), op, ns;
```

## Export to TSV File
Export to a TVS file and import to a spreadsheet software.  Here is an example"
```sqlite3
sqlite3 -header -separator " " ./data/hatchet.db "SELECT * FROM mongod_json;" > mongod_json.tsv
```

## Hatchet API
### Slow Op Patterns Stats
```
/api/hatchet/v1.0/tables/{table}/stats/slowops
```

#### Get Slow Op Patterns
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_json/stats/slowops"
```
#### Get Slow Op Patterns Order By avg_ms
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_json/stats/slowops?orderBy=avg_ms"
```

Valid orderBy values are:
- op
- ns
- count
- avg_ms
- max_ms
- total_ms
- reslen

### Slow Op Logs
```
/api/hatchet/v1.0/tables/{table}/logs/slowops
```

#### Get Slowest Ops
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_json/logs/slowops"
```

#### Get 100 Slowest Ops
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_json/logs/slowops?topN=100"
```

## View in Browser
```
/tables/{table}
```

### View Stats in Browser
URL to view stats of slow ops:
```
/tables/{table}/stats/slowops[?COLLSCAN=&orderBy=]
```

URL to view logs of slow ops:
```
/tables/{table}/logs/slowops
```

#### View Slow Ops Stats
```
http://localhost:3721/tables/mongod_json/stats/slowops
```

#### View COLLSCAN Only Stats
```
http://localhost:3721/tables/mongod_json/stats/slowops?COLLSCAN=true
```

#### View Slow Ops Stats Order by Count
```
http://localhost:3721/tables/mongod_json/stats/slowops?orderBy=count
```

#### View COLLSCAN Stats Order by total_ms
```
http://localhost:3721/tables/mongod_json/stats/slowops?COLLSCAN=true&orderBy=total_ms
```

#### View Slowest Logs
It will show 25 or less slowest operations.
```
http://localhost:3721/tables/mongod_json/logs/slowops
```

#### View Top 100 Slowest Logs
```
http://localhost:3721/tables/mongod_json/logs/slowops?topN=100
```

### Search Logs
To show all logs:
```
http://localhost:3721/tables/mongod_json/logs
```

#### Search by
Possible search parameters are:
- component
- context
- duration (begin_datetime,end_datetime)
- severity

Multiple parameters have an AND relationship in the where clause.  For example:
```
http://localhost:3721/tables/mongod_json/logs?component=NETWORK&severity=E&duration=2021-07-25T09:25:14,2021-07-25T09:26:00
```

### View Charts in Browser
```
http://localhost:3721/tables/mongod_json/charts/slowops
```

## Output in Legacy Format
```bash
./dist/hatchet -legacy testdata/mongod_json.log.gz > mongod_text.log
```

## Use SQLite3 API
All major programming languages are supported from different drivers, including Golang, NodeJS, Java, Python, and C#.
