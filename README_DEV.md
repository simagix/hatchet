# Hatchet Developer's Guide
The Hatchet tool provides many different ways to access data and there are:
- Query the SQLite database
- Output to a TSV file and import data into a spreadsheet
- Use the provided HTML reports and view in a browser
- Output the logs in the legacy format and processed by other tools
- Use SQLite API to access data from applications
- Use Hatchet API to import JSON data into applications

Note that there are a few indexes created during the logs processings.  But,
you can create additional indexes to support additional needs.

## Query SQLite Database
The default database is stored in the *data/hatchet.db* file.
```bash
sqlite3 ./data/hatchet.db
```

### Useful SQLite3 Commands
After a log file is processed, a table is created in the SQLite database.  The table name is 
part of the log file name.  A table name *mongod_v2* is extracted from a log file name of, for example, 
$HOME/Downloads/**mongod_v2**.log.gz.

### Query Data
```sqlite3
SELECT * from mongod_v2;
```

```sqlite3
SELECT date, severity, component, context, substr(message, 1, 60) message FROM mongod_v2;
```

```sqlite3
SELECT date, severity, message FROM mongod_v2 WHERE component = 'NETWORK';
```

### Query Performance Data
```sqlite3
SELECT op, COUNT(*) "count", ROUND(AVG(milli),1) avg_ms, MAX(milli) max_ms, SUM(milli) total_ms,
       ns, _index "index", SUM(reslen) "reslen", filter "query pattern"
    FROM mongod_v2 
    WHERE op != "" GROUP BY op, ns, filter ORDER BY avg_ms DESC;
```

## Export to TSV File
Export to a TVS file and import to a spreadsheet software.  Here is an example"
```sqlite3
sqlite3 -header -separator " " ./data/hatchet.db "SELECT * FROM mongod_v2;" > mongod_v2.tsv
```

## Hatchet API
### Slow Op Patterns Stats
```
/api/hatchet/v1.0/tables/{table}/stats/slowops
```

#### Get Slow Op Patterns
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_v2/stats/slowops"
```
#### Get Slow Op Patterns Order By avg_ms
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_v2/stats/slowops?orderBy=avg_ms"
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
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_v2/logs/slowops"
```

#### Get 100 Slowest Ops
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_v2/logs/slowops?topN=100"
```

## View Stats in Browser
URL to view stats of slow ops:
```
/tables/{table}/stats/slowops[?COLLSCAN=&orderBy=]
```

URL to view logs of slow ops:
```
/tables/{table}/logs/slowops
```

### View Slow Ops Stats
```
http://localhost:3721/tables/mongod_v2/stats/slowops
```

### View COLLSCAN Only Stats
```
http://localhost:3721/tables/mongod_v2/stats/slowops?COLLSCAN=true
```

### View Slow Ops Stats Order by Count
```
http://localhost:3721/tables/mongod_v2/stats/slowops?orderBy=count
```

### View COLLSCAN Stats Order by total_ms
```
http://localhost:3721/tables/mongod_v2/stats/slowops?COLLSCAN=true&orderBy=total_ms
```

### View Slowest Logs
It will show 25 or less slowest operations.
```
http://localhost:3721/tables/mongod_v2/logs/slowops
```

### View Top 100 Slowest Logs
```
http://localhost:3721/tables/mongod_v2/logs/slowops?topN=100
```

### Search Logs
To show all logs:
```
http://localhost:3721/tables/mongod_v2/logs
```

#### Search by
Possible search parameters are:
- component
- context
- duration (begin_datetime,end_datetime)
- severity

Multiple parameters have an AND relationship in the where clause.  For example:
```
http://localhost:3721/tables/mongod_v2/logs?component=NETWORK&severity=E&duration=2021-07-25T09:25:14,2021-07-25T09:26:00
```
