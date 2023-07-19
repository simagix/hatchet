# Hatchet - MongoDB JSON Log Analyzer and Viewer
Hatchet is a powerful and sophisticated logs analyzer and viewer specifically designed for MongoDB JSON logs. It provides advanced features for logs processing, aggregation and storage of the processed data. To make the data accessible and usable for its users, Hatchet utilizes an embedded SQLite3 database. This database allows for the storage of processed and aggregated data and makes it possible to offer RESTful APIs and a web interface to users.

The web interface of Hatchet is highly interactive and user-friendly, providing a seamless experience for searching logs and navigating through reports and charts. The intuitive design and easy-to-use interface makes it simple for users to find what they need, when they need it. Additionally, with the embedded database, Hatchet provides fast access to data and a high level of performance, making it the ideal solution for logs analysis and management. Further design details can be found at [Streamline Your MongoDB Log Analysis with Hatchet](https://www.simagix.com/2023/02/streamline-your-mongodb-log-analysis.html) and [Hatchet: Empowering Smart MongoDB Log Analysis 
](https://www.simagix.com/2023/07/hatchet-empowering-smart-mongodb-log.html).

## Build
Clone and run the *build.sh* script; *gcc* is required to support CGO.
```bash
git clone --depth 1 git@github.com:simagix/hatchet.git
cd hatchet ; ./build.sh
```

An executable *hatchet* is output to the directory *dist/*.  Note that the script also works and tested on Windows x64 using MingGW and Git Bash.

## Quick Start
Use the command below to process a log file, mongod.log.gz and start a web server listening to port 3721.  The default database is SQLite3.
```bash
./dist/hatchet -web mongod.log.gz
```

Use the URL `http://localhost:3721/` in a browser to view reports and charts.  Alternatively, you can use the *in-memory* mode without persisting data, for example:
```bash
./dist/hatchet -url in-memory mongod.log.gz
```

Using MongoDB as the database is also supported.  The example below stores data in collections under *logdb* database.
```bash
./dist/hatchet -url mongodb+srv://{user}:{password}@demo.host.mongodb.net/logdb mongod.log.gz
```

if you choose to view in the legacy format without a browser, use the command below:
```bash
./dist/hatchet -legacy mongod.log.gz
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
