// Copyright 2022-present Kuei-chun Chen. All rights reserved.
package hatchet

type Database interface {
	Begin() error
	Close() error
	Commit() error

	GetAcceptedConnsCounts(tableName string, duration string) ([]NameValue, error)
	GetAverageOpTime(tableName string, duration string) ([]OpCount, error)
	GetClientPreparedStmt(id string) string
	GetConnectionStats(tableName string, chartType string, duration string) ([]Remote, error)
	GetHatchetInitStmt(tableName string) string
	GetHatchetPreparedStmt(id string) string
	GetLogs(tableName string, opts ...string) ([]LegacyLog, error)
	GetLogsFromMessage(tableName string, opts ...string) ([]LegacyLog, error)
	GetOpsCounts(tableName string, duration string) ([]NameValue, error)
	GetSlowOps(tableName string, orderBy string, order string, collscan bool) ([]OpStat, error)
	GetSlowOpsCounts(tableName string, duration string) ([]OpCount, error)
	GetSlowOpsQuery(tableName string, orderBy string, order string, collscan bool) string
	GetSlowestLogs(tableName string, topN int) ([]string, error)
	GetSubStringFromTable(tableName string) string
	GetTableSummary(tableName string) (string, string, string)
	GetTables() ([]string, error)
	GetVerbose() bool

	InsertClientConn(index int, rmt Remote) error
	InsertLog(index int, end string, doc *Logv2Info, stat *OpStat) error
	InsertOps() error
	UpdateHatchetInfo(info HatchetInfo) error

	SetVerbose(v bool)
}

func GetDatabase() (Database, error) {
	var dbase Database
	var err error
	logv2 := GetLogv2()
	if dbase, err = NewSQLite3DB(logv2.dbfile, logv2.tableName); err != nil {
		return dbase, err
	}
	dbase.SetVerbose(logv2.verbose)
	return dbase, err
}
