// Copyright 2022-present Kuei-chun Chen. All rights reserved.

package hatchet

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/simagix/gox"
)

const SQLITE3_FILE = "./data/hatchet.db"

func Run(fullVersion string) {
	dbfile := flag.String("dbfile", SQLITE3_FILE, "database file name")
	legacy := flag.Bool("legacy", false, "view logs in legacy format")
	port := flag.Int("port", 3721, "web server port number")
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("verbose", false, "turn on verbose")
	web := flag.Bool("web", false, "test mode")
	flag.Parse()
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if *ver {
		fmt.Println(fullVersion)
		return
	}

	logv2 := Logv2{dbfile: *dbfile, verbose: *verbose, legacy: *legacy}
	instance = &logv2
	for _, filename := range flag.Args() {
		err := logv2.Analyze(filename)
		if err != nil {
			log.Fatal(err)
		}
	}
	if *legacy || !*web {
		return
	}

	http.HandleFunc("/", gox.Cors(handler))
	http.HandleFunc("/api/hatchet/v1.0/tables/", gox.Cors(apiHandler))
	http.HandleFunc("/tables/", gox.Cors(htmlHandler))
	addr := fmt.Sprintf(":%d", *port)
	if listener, err := net.Listen("tcp", addr); err != nil {
		log.Fatal(err)
	} else {
		listener.Close()
		log.Println("starting web server", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

