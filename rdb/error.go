package rdb

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"strings"
)

const (
	mysqlDuplicateErrNo        = 1062
	mysqlExecInReadOnlyTxErrNo = 1792
)

var (
	ErrNotFound         = errors.New("not found")
	ErrNothingUpdated   = errors.New("nothing updated")
	ErrDuplicateKey     = errors.New("duplicate key")
	ErrAffectMany       = errors.New("affect many rows")
	ErrExecInReadOnlyTx = errors.New("cannot exec in read-only transaction")
)

func IsMySQLDuplicateError(err error) bool {
	dr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return dr.Number == mysqlDuplicateErrNo
}

func IsMySQLExecInReadOnlyTxError(err error) bool {
	dr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return dr.Number == mysqlExecInReadOnlyTxErrNo
}

func IsMySQLNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "no rows")
}
