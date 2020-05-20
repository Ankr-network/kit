package rdb

import (
	"database/sql/driver"
	"errors"
	"time"
)

const (
	zeroUnixNanos = -6795364578871345152
)

var (
	ErrInvalidDBValueForTime = errors.New("invalid db value for time")
)

type Time struct {
	time time.Time
}

func FromTime(t time.Time) Time {
	return Time{time: t}
}

func (t *Time) ToTime() time.Time {
	return t.time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return t.time.MarshalJSON()
}

func (t Time) String() string {
	return t.time.String()
}

// Scan implements the Scanner interface.
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	int64V, int64OK := value.(int64)
	if !int64OK {
		return ErrInvalidDBValueForTime
	}

	if int64V == zeroUnixNanos {
		return nil
	}

	t.time = time.Unix(0, int64V)
	return nil
}

// Value implements the driver Valuer interface.
func (t Time) Value() (driver.Value, error) {
	return t.time.UnixNano(), nil
}

func (t Time) Equal(ct Time) bool {
	return t.time.Equal(t.time)
}
