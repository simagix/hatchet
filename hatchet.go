/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * hatchet.go
 */

package hatchet

import (
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"

	"github.com/julienschmidt/httprouter"
	"github.com/mattn/go-sqlite3"
)

const SQLITE3_FILE = "./data/hatchet.db"

func Run(fullVersion string) {
	dbfile := flag.String("dbfile", SQLITE3_FILE, "database file name")
	digest := flag.Bool("digest", false, "HTTP digest")
	endpoint := flag.String("endpoint-url", "", "AWS endpoint")
	inMem := flag.Bool("in-memory", false, "use in-memory mode")
	legacy := flag.Bool("legacy", false, "view logs in legacy format")
	port := flag.Int("port", 3721, "web server port number")
	profile := flag.String("aws-profile", "default", "AWS profile name")
	s3 := flag.Bool("s3", false, "files from AWS S3")
	user := flag.String("user", "", "HTTP Auth (username:password)")
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("verbose", false, "turn on verbose")
	web := flag.Bool("web", false, "starts a web server")
	flag.Parse()
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if *ver {
		fmt.Println(fullVersion)
		return
	}
	if !*legacy {
		log.Println(fullVersion)
	}

	if *inMem {
		if *dbfile != SQLITE3_FILE {
			log.Fatalln("cannot use -dbfile and -in-memory together")
		} else if len(flag.Args()) == 0 {
			log.Fatalln("cannot use -in-memory without a log file")
		}
		log.Println("in-memory mode is enabled, no data will be persisted")
		*dbfile = "file::memory:?cache=shared"
		*web = true
	}

	regex := func(re, s string) (bool, error) {
		return regexp.MatchString(re, s)
	}
	sql.Register("sqlite3_extended",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regex, true)
			},
		})
	logv2 := Logv2{version: fullVersion, dbfile: *dbfile, verbose: *verbose, legacy: *legacy, user: *user, isDigest: *digest}
	if *s3 {
		var err error
		if logv2.s3client, err = NewS3Client(*profile, *endpoint); err != nil {
			log.Fatal(err)
		}
	}
	instance = &logv2
	for _, filename := range flag.Args() {
		err := logv2.Analyze(filename)
		if err != nil {
			log.Fatal(err)
		}
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

	router.GET("/api/hatchet/v1.0/hatchets/:hatchet/:category/:attr", APIHandler)

	router.GET("/hatchets/:hatchet/charts/:attr", ChartsHandler)
	router.GET("/hatchets/:hatchet/logs/:attr", LogsHandler)
	router.GET("/hatchets/:hatchet/stats/:attr", StatsHandler)

	addr := fmt.Sprintf(":%d", *port)
	if listener, err := net.Listen("tcp", addr); err != nil {
		log.Fatal(err)
	} else {
		listener.Close()
		if !*inMem {
			log.Println("using data file", *dbfile)
		}
		log.Println("starting web server at", addr)
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
