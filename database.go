/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * database.go
 */

package hatchet

import "log"

const (
	SQLite3 = iota
	Mongo
)

type NameValue struct {
	Name  string `bson:"name"`
	Value int    `bson:"value"`
}

type NameValues struct {
	Name   string
	Values []interface{}
}

type Database interface {
	Begin() error
	Close() error
	Commit() error
	CountLogs(opts ...string) (int, error)
	CreateMetaData() error
	Drop() error
	GetAcceptedConnsCounts(duration string) ([]NameValue, error)
	GetAuditData() (map[string][]NameValues, error)
	GetAverageOpTime(op string, duration string) ([]OpCount, error)
	GetConnectionStats(chartType string, duration string) ([]RemoteClient, error)
	GetHatchetInfo() HatchetInfo
	GetHatchetNames() ([]string, error)
	GetLogs(opts ...string) ([]LegacyLog, error)
	GetOpsCounts(duration string) ([]NameValue, error)
	GetReslenByAppName(appname string, duration string) ([]NameValue, error)
	GetReslenByNamespace(ip string, duration string) ([]NameValue, error)
	GetReslenByIP(ip string, duration string) ([]NameValue, error)
	GetSlowOps(orderBy string, order string, collscan bool) ([]OpStat, error)
	GetSlowestLogs(topN int) ([]LegacyLog, error)
	GetVerbose() bool
	InsertClientConn(index int, doc *Logv2Info) error
	InsertDriver(index int, doc *Logv2Info) error
	InsertFailedMessages(m *FailedMessages) error
	InsertLog(index int, end string, doc *Logv2Info, stat *OpStat) error
	SearchLogs(opts ...string) ([]LegacyLog, error)
	SetVerbose(v bool)
	UpdateHatchetInfo(info HatchetInfo) error
}

func GetDatabase(hatchetName string) (Database, error) {
	var err error
	var dbase Database
	logv2 := GetLogv2()
	if logv2.verbose {
		log.Println("url", logv2.url, "hatchet name", hatchetName)
	}
	if GetLogv2().GetDBType() == Mongo {
		if dbase, err = NewMongoDB(logv2.url, hatchetName); err != nil {
			return nil, err
		}
	} else { // default is SQLite3
		if dbase, err = NewSQLite3DB(logv2.url, hatchetName, logv2.cacheSize); err != nil {
			return nil, err
		}
	}
	dbase.SetVerbose(logv2.verbose)
	return dbase, err
}

// GetExistingHatchetNames returns a list of existing hatchet names from the database
func GetExistingHatchetNames() ([]string, error) {
	var err error
	var dbase Database
	logv2 := GetLogv2()
	// Use a temporary name just to connect and query
	if GetLogv2().GetDBType() == Mongo {
		if dbase, err = NewMongoDB(logv2.url, "_temp"); err != nil {
			return nil, err
		}
	} else {
		if dbase, err = NewSQLite3DB(logv2.url, "_temp", logv2.cacheSize); err != nil {
			return nil, err
		}
	}
	defer dbase.Close()
	return dbase.GetHatchetNames()
}
