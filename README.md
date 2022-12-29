# Hatchet - MongoDB logv2 Analyzer
Hatchet is a MongoDB logv2 analyzer and it stores parsed data in the embbedded SQLite3 database.

## Build
Clone and run the *build.sh* script.  For the first installation, *gcc* is required to support CGO.
```bash
./build.sh
```

## Usages
Hatchet can be used in multiple ways and they are:
- Parse logs and store in SQLite3 database
- View logs in legacy format using *-legacy* flag
- Start as a web server to view HTML reports

### Process Log File(s)
Use the command below to process log files and obtain hatchet IDs.

```bash
./dist/hatchet <log file> [<more log file>...]
```

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

### View in Browser
####View Slow Op Stats
```
/tables/{table}/stats/slowops
```

####View Slow Op Stats Logs
```
/tables/{table}/logs/slowops
```

See [developer's guide](README_DEV.md) for Hatchet integration  details.

## License
[Apache-2.0 License](LICENSE)
