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
Start a web server with *-web* option.  The default port is 3721.
```bash
./dist/hatchet -web [<log file>...]
```

### View from Browser
**View Slow Op Stats Summary**
```
/tables/{table}/slowops/summary
```

For example `http://localhost:3721/tables/mongod_v2/slowops/summary`

**View Slow Op Stats Summary Order by count**
```
/tables/{table}/slowops/summary?orderBy=count
```

For example `http://localhost:3721/tables/mongod_v2/slowops/summary?orderBy=count`

For Hatchet APIs, see [developer's guide](README_DEV.md) for more details.

## License
[Apache-2.0 License](LICENSE)
