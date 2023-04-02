/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * database.go
 */

package hatchet

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
	GetReslenByNamespace(ip string, duration string) ([]NameValue, error)
	GetReslenByIP(ip string, duration string) ([]NameValue, error)
	GetSlowOps(orderBy string, order string, collscan bool) ([]OpStat, error)
	GetSlowestLogs(topN int) ([]LegacyLog, error)
	GetVerbose() bool
	InsertClientConn(index int, doc *Logv2Info) error
	InsertDriver(index int, doc *Logv2Info) error
	InsertLog(index int, end string, doc *Logv2Info, stat *OpStat) error
	SearchLogs(opts ...string) ([]LegacyLog, error)
	SetVerbose(v bool)
	UpdateHatchetInfo(info HatchetInfo) error
}

func GetDatabase(hatchetName string) (Database, error) {
	if GetLogv2().GetDBType() == Mongo {
		return GetMongoDB(hatchetName)
	}
	return GetSQLite3DB(hatchetName)
}
