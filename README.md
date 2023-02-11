# Hatchet - MongoDB JSON Log Analyzer and Viewer
Hatchet is a MongoDB JSON logs analyzer and viewer.  It stores processed and aggregated data in an embedded SQLite3 database to support RESTful APIs and a web interface.  With an embedded database, Hatchet provides an interactive users experience to search logs and to navigate reports and charts.  See [Streamline Your MongoDB Log Analysis with Hatchet](https://www.simagix.com/2023/02/streamline-your-mongodb-log-analysis.html) for more details.

## Build
Clone and run the *build.sh* script; *gcc* is required to support CGO.
```bash
git clone --depth 1 git@github.com:simagix/hatchet.git
cd hatchet ; ./build.sh
```

An executable *hatchet* is output to the directory *dist/*.

## Quick Start
Use the command below to process a log file, mongod.log.gz and start a web server listening to port 3721.
```bash
./dist/hatchet -web mongod.log.gz
```

Use the URL `http://localhost:3721/` in a browser to view reports and charts.  Alternatively, you can use the in-memory mode without persisting data, for example:
```bash
./dist/hatchet -in-memory mongod.log.gz
```

if you choose to view in the legacy format without a browser, use the command below:
```bash
./dist/hatchet -legacy mongod.log.gz
```

For additional usages and integration details, see [developer's guide](README_DEV.md).

## License
[Apache-2.0 License](LICENSE)
