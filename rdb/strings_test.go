package rdb

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStructToString(t *testing.T) {
	a := Strings{}
	err := a.Scan([]byte("1,2,3,4"))
	assert.NoError(t, err)
	assert.Equal(t, a, Strings([]string{"1", "2", "3", "4"}))
	value, err := a.Value()
	assert.NoError(t, err)
	t.Log(value)

	err = a.Scan(123123)
	assert.True(t, errors.Is(err, ErrInvalidDBValueForStructToString))
}
