package mdb

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DuplicateErrorCode = 11000
)

var (
	ErrNotFound     = errors.New("not found")
	ErrDuplicateKey = errors.New("duplicate key")
)

func IsDuplicateKeyError(err error) bool {
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		return cmdErr.Code == DuplicateErrorCode
	}
	var writeEx mongo.WriteException
	if errors.As(err, &writeEx) {
		if len(writeEx.WriteErrors) == 0 {
			return false
		}
		for _, we := range writeEx.WriteErrors {
			if we.Code == DuplicateErrorCode {
				return true
			}
		}
		return false
	}
	return false
}
