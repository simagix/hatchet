# Hatchet Developer's Guide

## Query Database
The default database is stored in the *data/hatchet.db* file.
```bash
sqlite3 ./data/hatchet.db
```

### Useful SQLite3 Commands
```sqlite3
.header on
.mode column
.tables
.schema
```

### Query All Data
```sqlite3
SELECT * from mongod_v2;

SELECT date, severity, component, context, substr(message, 1, 60) message 
    FROM mongod_v2;
```

### Query by Component and Context
```sqlite3
SELECT date, severity, message 
    FROM mongod_v2 
    WHERE component = 'NETWORK' AND context = 'listener';
```

### Query Performance Data
```sqlite3
SELECT op, COUNT(*) "count", ROUND(AVG(milli),1) avg_ms, MAX(milli) max_ms, SUM(milli) total_ms,
       ns, _index "index", SUM(reslen) "reslen", filter "query pattern"
    FROM mongod_v2 
    WHERE op != "" 
    GROUP BY op, ns, filter
    ORDER BY avg_ms DESC;
```

## Hatchet API
### Slow Op Patterns Summary
```
/api/hatchet/v1.0/tables/{table}/slowops/summary
```

**Get Slow Op Patterns**
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_v2/slowops/summary"
```
**Get Slow Op Patterns Order By avg_ms**
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_v2/slowops/summary?orderBy=avg_ms"
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
/api/hatchet/v1.0/tables/{table}/slowops/logs
```

**Get Top 25 Slowest Ops**
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_v2/slowops/logs"
```
**Get Top 100 Slowest Ops**
```bash
curl "http://localhost:3721/api/hatchet/v1.0/tables/mongod_v2/slowops/logs?topN=100"
```