//+build integration

package auth

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisBlacklist(t *testing.T) {
	bl := NewRedisBlacklist(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	}))

	bl.PutAccess("test1", time.Now(), 1*time.Second)
	assert.Error(t, bl.CheckAccess("test1"), ErrExpiredAccess)

	bl.PutAccess("test2", time.Now().Add(-2*time.Second), 1*time.Second)
	assert.NoError(t, bl.CheckAccess("test2"))
}
