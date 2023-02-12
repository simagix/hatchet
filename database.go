// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

import "log"

type Database interface {
	Begin() error
	Close() error
	Commit() error
	GetAcceptedConnsCounts(duration string) ([]NameValue, error)
	GetAverageOpTime(duration string) ([]OpCount, error)
	GetClientPreparedStmt() string
	GetConnectionStats(chartType string, duration string) ([]Remote, error)
	GetHatchetInfo() HatchetInfo
	GetHatchetInitStmt() string
	GetHatchetNames() ([]string, error)
	GetHatchetPreparedStmt() string
	GetLogs(opts ...string) ([]LegacyLog, error)
	GetOpsCounts(duration string) ([]NameValue, error)
	GetReslenByClients(duration string) ([]NameValue, error)
	GetSlowOps(orderBy string, order string, collscan bool) ([]OpStat, error)
	GetSlowOpsCounts(duration string) ([]OpCount, error)
	GetSlowestLogs(topN int) ([]LegacyLog, error)
	GetVerbose() bool
	InsertClientConn(index int, doc *Logv2Info) error
	InsertLog(index int, end string, doc *Logv2Info, stat *OpStat) error
	InsertOps() error
	SearchLogs(opts ...string) ([]LegacyLog, error)
	SetVerbose(v bool)
	UpdateHatchetInfo(info HatchetInfo) error
}

func GetDatabase(hatchetName string) (Database, error) {
	var dbase Database
	var err error
	logv2 := GetLogv2()
	if logv2.verbose {
		log.Println("dbfile", logv2.dbfile, "hatchet name", hatchetName)
	}
	if dbase, err = NewSQLite3DB(logv2.dbfile, hatchetName); err != nil {
		return dbase, err
	}
	dbase.SetVerbose(logv2.verbose)
	return dbase, err
}
