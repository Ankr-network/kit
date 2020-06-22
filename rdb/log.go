package rdb

import (
	"kit/mlog"
	"strings"
)

var (
	log = mlog.Logger("rdb")
)

func LogSQL(sql string, args ...interface{}) {
	outSQL := strings.ReplaceAll(sql, "?", "%v")
	log.Sugar().Debugf(outSQL, args...)
}
