package rdb

import (
	"github.com/Ankr-network/kit/mlog"
	"strings"
)

var (
	log = mlog.Logger("pkg.rdb")
)

func LogSQL(sql string, args ...interface{}) {
	outSQL := strings.ReplaceAll(sql, "?", "%v")
	log.Sugar().Debugf(outSQL, args...)
}
