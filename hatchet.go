/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * hatchet.go
 */

package hatchet

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mattn/go-sqlite3"
)

const SQLITE3_FILE = "./data/hatchet.db"

func Run(fullVersion string) {
	bios := flag.Bool("bios", false, "populate bios documents")
	cache := flag.Int("cache_size", 2000, "number of cache pages")
	connstr := flag.String("url", SQLITE3_FILE, "database file name or connection string")
	digest := flag.Bool("digest", false, "HTTP digest")
	endpoint := flag.String("endpoint-url", "", "AWS endpoint")
	from := flag.String("from", "1970-01-01T00:00:00Z", "from date/time")
	merge := flag.Bool("merge", false, "merge files")
	legacy := flag.Bool("legacy", false, "view logs in legacy format")
	infile := flag.String("obfuscate", "", "obfuscate logs")
	port := flag.Int("port", 3721, "web server port number")
	profile := flag.String("aws-profile", "default", "AWS profile name")
	s3 := flag.Bool("s3", false, "files from AWS S3")
	sim := flag.String("sim", "", "simulate read/write load tests")
	to := flag.String("to", "", "from date/time")
	user := flag.String("user", "", "HTTP Auth (username:password)")
	verbose := flag.Bool("v", false, "turn on verbose")
	ver := flag.Bool("version", false, "print version number")
	web := flag.Bool("web", false, "starts a web server")
	server := flag.Bool("server", false, "starts a web server (alias for -web)")
	flag.Parse()
	// -server is an alias for -web
	if *server {
		*web = true
	}
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if *ver {
		fmt.Println(fullVersion)
		return
	} else if *infile != "" {
		// Create obfuscated directory
		outDir := "obfuscated"
		if err := os.MkdirAll(outDir, 0755); err != nil {
			log.Fatal(err)
		}

		// Determine output filenames
		baseName := filepath.Base(*infile)
		if baseName == "-" {
			baseName = "stdin"
		}
		// Remove compression extensions for cleaner output names
		baseName = strings.TrimSuffix(baseName, ".gz")
		baseName = strings.TrimSuffix(baseName, ".zst")
		logFile := filepath.Join(outDir, baseName)
		mapFile := filepath.Join(outDir, baseName+".mappings.json")

		// Create output log file
		outFile, err := os.Create(logFile)
		if err != nil {
			log.Fatal(err)
		}

		obs := NewObfuscation()
		err = obs.ObfuscateFileToWriter(*infile, outFile)
		outFile.Close()
		if err != nil {
			log.Fatal(err)
		}

		// Save mappings to JSON file
		var data []byte
		if data, err = json.MarshalIndent(*obs, "", "  "); err != nil {
			log.Fatal(err)
		}
		if err = os.WriteFile(mapFile, data, 0644); err != nil {
			log.Fatal(err)
		}
		log.Printf("Obfuscated log saved to %s\n", logFile)
		log.Printf("Mappings saved to %s\n", mapFile)
		return
	} else if *bios && len(flag.Args()) > 1 {
		InsertBiosIntoMongoDB(flag.Args()[0], ToInt(flag.Args()[1]))
		return
	} else if *sim != "" && len(flag.Args()) > 0 {
		SimulateTests(*sim, flag.Args()[0])
		return
	}
	if !*legacy {
		log.Println(fullVersion)
	}

	if *connstr == "in-memory" {
		if len(flag.Args()) == 0 {
			log.Fatalln("cannot use -in-memory without a log file")
		}
		log.Println("in-memory mode is enabled, no data will be persisted")
		*connstr = "file::memory:?cache=shared"
		*web = true
	}

	layout := "2006-01-02T15:04:05"
	fromTime, err := time.Parse(layout, *from)
	if err != nil {
		fromTime, err = time.Parse(layout, "2000-01-01T00:00:00")
	}
	toTime, err := time.Parse(layout, *to)
	if err != nil {
		toTime = time.Now()
	}
	logv2 := Logv2{version: fullVersion, url: *connstr, verbose: *verbose,
		legacy: *legacy, user: *user, isDigest: *digest, cacheSize: *cache,
		from: fromTime, to: toTime, merge: *merge}
	if *merge {
		logv2.hatchetName = getHatchetName("merge")
	}
	instance = &logv2
	str := *connstr
	if strings.HasPrefix(*connstr, "mongodb") {
		log.Fatalln("MongoDB backend is no longer supported. Please use SQLite3 (default) or in-memory mode.")
	}
	log.Println("using database", str)
	if GetLogv2().GetDBType() == SQLite3 {
		regex := func(re, s string) (bool, error) {
			return regexp.MatchString(re, s)
		}
		sql.Register("sqlite3_extended",
			&sqlite3.SQLiteDriver{
				ConnectHook: func(conn *sqlite3.SQLiteConn) error {
					return conn.RegisterFunc("regexp", regex, true)
				},
			})
	}
	if *s3 {
		var err error
		if logv2.s3client, err = NewS3Client(*profile, *endpoint); err != nil {
			log.Fatal(err)
		}
	}
	for i, logname := range flag.Args() {
		if err := logv2.Analyze(logname, i+1); err != nil {
			log.Fatal(err)
		}
		if !*merge && !*legacy {
			logv2.PrintSummary()
		}
	}
	if *merge && !*legacy {
		logv2.PrintSummary()
	}
	if *legacy || !*web {
		if len(flag.Args()) == 0 {
			flag.PrintDefaults()
		}
		return
	}

	router := httprouter.New()
	router.GET("/", Handler)
	router.GET("/favicon.ico", FaviconHandler)

	router.GET("/api/hatchet/v1.0/mongodb/:mongo/drivers/:driver", DriverHandler)
	router.GET("/api/hatchet/v1.0/hatchets/:hatchet/:category/:attr", APIHandler)

	router.GET("/hatchets/:hatchet/charts/:attr", ChartsHandler)
	router.GET("/hatchets/:hatchet/logs/:attr", LogsHandler)
	router.GET("/hatchets/:hatchet/stats/:attr", StatsHandler)
	router.POST("/api/hatchet/v1.0/rename", RenameHandler)
	router.DELETE("/api/hatchet/v1.0/delete", DeleteHandler)
	router.POST("/api/hatchet/v1.0/upload", UploadHandler)
	router.GET("/api/hatchet/v1.0/upload/status/:name", UploadStatusHandler)

	addr := fmt.Sprintf(":%d", *port)
	if listener, err := net.Listen("tcp", addr); err != nil {
		log.Fatal(err)
	} else {
		listener.Close()
		log.Println("starting web server at", addr)
		log.Printf("http://localhost:%d\n", *port)
		log.Fatal(http.ListenAndServe(addr, router))
	}
}

func FaviconHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	r.Close = true
	r.Header.Set("Connection", "close")
	w.Header().Set("Content-Type", "image/x-icon")
	ico, _ := base64.StdEncoding.DecodeString(CHEN_ICO)
	w.Write(ico)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}
