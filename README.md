# Hatchet - MongoDB JSON Log Analyzer
Hatchet is a MongoDB JSON logs analyzer and it stores processed data in a embbedded SQLite3 database.

## Build
Clone and run the *build.sh* script; *gcc* is required to support CGO.
```bash
./build.sh
```

An executable *hatchet* is output to the directory *dist/*.

## Quick Start
Use the command below to process a log file, mongod.log.gz and start a web server listening to port 3721.
```bash
./dist/hatchet -web mongod.log.gz
```

Use the URL `http://localhost:3721/` in a browser to view reports and charts.  if you choose to view in the
legacy format without a browser, use the command below:
```bash
./dist/hatchet -legacy mongod.log.gz
```

For additional usages and integration details, see [developer's guide](README_DEV.md).

## License
[Apache-2.0 License](LICENSE)
