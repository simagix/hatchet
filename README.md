# Hatchet - MongoDB JSON Log Analyzer and Viewer
Hatchet is a powerful and sophisticated logs analyzer and viewer specifically designed for MongoDB JSON logs. It provides advanced features for logs processing, aggregation and storage of the processed data. To make the data accessible and usable for its users, Hatchet utilizes an embedded SQLite3 database. This database allows for the storage of processed and aggregated data and makes it possible to offer RESTful APIs and a web interface to users.

[![Github sponsor](https://img.shields.io/badge/sponsor-30363D?style=for-the-badge&logo=GitHub-Sponsors&logoColor=#white)](https://github.com/sponsors/simagix)

The web interface of Hatchet is highly interactive and user-friendly, providing a seamless experience for searching logs and navigating through reports and charts. The intuitive design and easy-to-use interface makes it simple for users to find what they need, when they need it. Additionally, with the embedded database, Hatchet provides fast access to data and a high level of performance, making it the ideal solution for logs analysis and management. Further design details can be found at [![Hatchet: Empowering Smart MongoDB Log Analysis](http://img.youtube.com/vi/WavOyaFTDE8/0.jpg)](https://www.youtube.com/watch?v=WavOyaFTDE8).

## Change Log
- [v0.5.0](https://youtu.be/4RkeMOOAtv8), August, 2023

## Build
Clone and run the *build.sh* script; *gcc* is required to support CGO.
```bash
git clone --depth 1 https://github.com/simagix/hatchet.git
cd hatchet ; ./build.sh
```

An executable *hatchet* is output to the directory *dist/*.  Note that the script also works and tested on Windows x64 using MingGW and Git Bash.

## Quick Start
Use the command below to process a log file, mongod.log.gz and start a web server listening to port 3721.  The default database is SQLite3.
```bash
./dist/hatchet -web logs/sample-mongod.log.gz
```

## Docker
If you wish to use it via docker you can run:
```
# this will look after the build and start container for you
docker compose up -d 

# this will give you a bash shell on the container so that you can run commands
docker exec -it hatchet_container /bin/sh 

in all commands you can just use `hatchet` instead of `./dist/hatchet` for example

`hatchet -web logs/replica.tar.gz`

any files in the logs folder on your system will be available in the container

you can access http://localhost:3721 to view the hatchet web UI.

to clean up run `docker compose down` it will stop the hatchet container and delete it
```

Load a file within a defined time range:
```bash
./dist/hatchet -web -from "2023-09-23T20:25:00" -to "2023-09-23T20:26:00" logs/sample-mongod.log.gz
```

Load multiple files and process them individually:
```bash
./dist/hatchet -web rs1/mongod.log rs2/mongod.log rs3/mongod.log
```

Load multiple files and process them collectively:
```bash
./dist/hatchet -web -merge rs1/mongod.log rs2/mongod.log rs3/mongod.log
```

Use the URL `http://localhost:3721/` in a browser to view reports and charts.  Alternatively, you can use the *in-memory* mode without persisting data, for example:
```bash
./dist/hatchet -url in-memory logs/sample-mongod.log.gz
```

## Web UI Features
When running as a web service, Hatchet provides a rich set of features through its web interface:

### Upload Log Files
Upload MongoDB log files directly through the web interface - no command line needed. Simply drag and drop files onto the upload zone or click to browse. Supports any MongoDB log file (including `.gz` compressed). Multiple concurrent uploads are supported.

### Share Analysis via Direct Links
Share your analysis with team members using direct URLs:
- `/hatchets/{name}/stats/audit` - Security audit report
- `/hatchets/{name}/stats/slowops` - Slow query statistics
- `/hatchets/{name}/charts/operations` - Performance charts

### Download Reports
Download Audit and Stats reports as standalone HTML files for offline viewing or sharing via email/Slack. Click the "Download" button on any report page.

### Manage Hatcheted Logs
- **Rename**: Click the pencil icon to rename a hatcheted log
- **Delete**: Click the trash icon to remove a hatcheted log

### REST API
Hatchet provides a REST API for programmatic access:
- `POST /api/hatchet/v1.0/upload` - Upload log file (multipart form)
- `GET /api/hatchet/v1.0/upload/status/{name}` - Check upload status
- `POST /api/hatchet/v1.0/rename?old={name}&new={name}` - Rename hatchet
- `DELETE /api/hatchet/v1.0/delete?name={name}` - Delete hatchet
- `GET /api/hatchet/v1.0/hatchets/{name}/stats/audit` - Get audit data (JSON)
- `GET /api/hatchet/v1.0/hatchets/{name}/stats/slowops` - Get slow ops data (JSON)

if you choose to view in the legacy format without a browser, use the command below:
```bash
./dist/hatchet -legacy logs/sample-mongod.log.gz
```

For additional usages and integration details, see [developer's guide](README_DEV.md).

## A Smart Log Analyzer
How smart Hatchet is?  A picture is worth a thousand words.

![Sage Says](sage_says.png)

## Other Usages
Other than its ability to read from files, Hatchet offers additional functionality that includes reading from S3 and web servers, as well as MongoDB Atlas. This means that users can use Hatchet to conveniently access and download data from these sources, providing a more versatile and efficient data analysis experience.

### Web Servers
The tool supports reading from web servers using both the *http://* and *https://* protocols. The `-user` flag is optional when using basic authentication.

```bash
hatchet [-user {username}:{password}] https://{hostname}/{log name}
```

### Atlas
To download logs directly from MongoDB Atlas, you will need to use the `-user` and `-digest` flags and provide the necessary information for both. These flags are used to authenticate and authorize your access to the database.

```bash
hatchet -user {pub key}:{private key} -digest https://cloud.mongodb.com/api/atlas/v1.0/groups/{group ID}/clusters/{hostname}/logs/mongodb.gz
```

### AWS S3
Hatchet has the ability to download files from AWS S3. When downloading files, Hatchet will automatically retrieve the *Region* and *Credentials* information from the configuration files located at *${HOME}/.aws*. This means that there's no need to provide this information manually each time you download files from AWS S3 using Hatchet.

```bash
hatchet -s3 [--endpoint-url {test endpoint}] {bucket}/{key name}
```

## Logs Obfuscation
Use Hatchet to obfuscate logs. It automatically obfuscates the values of the matched patterns under the "attr" field, such as SSN, credit card numbers, phone numbers, email addresses, IP addresses, FQDNs, port numbers, namespaces, and other numbers. Note that, for example, replacing "host.example.com" with "rose.taipei.com" in the log file will consistently replace all other occurrences of "host.example.com" with "rose.taipei.com". To obfuscate logs and redirect them to a file, use the following syntax:

```bash
hatchet -obfuscate {log file} > {output file}
```

## License
[Apache-2.0 License](LICENSE)
