# MongoDB logv2 Analyzer
Hatchet is a MongoDB logv2 analyzer and it stores parsed data in the embbedded SQLite database.

## Build
Clone and run the *build.sh* script.  For the first installation, *gcc* is required to support CGO.
```bash
./build.sh
```

## Usages
Hatchet can be used in multiple ways and they are:
- Parse logs and store in SQLite database
- View logs in legacy format using *-legacy* flag
- Start as a web server to view HTML reports

### Process Log File(s)
Use the command below to process log files and obtain hatchet IDs.

```bash
./dist/hatchet <log file> [<more log file>...]
```
After a log file is processed, a table is created in the SQLite database.  The table name is 
part of the log file name.  A table name *mongod_v2* is extracted from a log file name of, for example, 
$HOME/Downloads/**mongod_v2**.log.gz.

### View Logs in Legacy Format
Use with *-legacy* and it prints logs to stdout.
```bash
./dist/hatchet -legacy <log file> [<more log file>...]
```

### Start a Web Server
Start a web server with *-web* option.
```bash
./dist/hatchet -web [<log file>...]
```
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

## License
[Apache-2.0 License](LICENSE)
