package auth

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

const (
	DefaultRedisBlacklistPrefix = "token:blacklist:"
)

var (
	ErrExpiredAccess = errors.New("expired or blocked access token")
)

type Blacklist interface {
	PutAccess(access string, createTime time.Time, expiration time.Duration) error
	CheckAccess(access string) error
}

type RedisBlacklistOptions struct {
	Prefix string
}

type RedisBlacklistOption func(opts *RedisBlacklistOptions)

type redisBlacklist struct {
	prefix string
	cli    redis.Cmdable
}

func WithPrefix(prefix string) RedisBlacklistOption {
	return func(opts *RedisBlacklistOptions) {
		opts.Prefix = prefix
	}
}

func NewRedisBlacklist(cmdable redis.Cmdable, opts ...RedisBlacklistOption) Blacklist {
	options := &RedisBlacklistOptions{
		Prefix: DefaultRedisBlacklistPrefix,
	}
	for _, opt := range opts {
		opt(options)
	}

	return &redisBlacklist{prefix: options.Prefix, cli: cmdable}
}

func (p *redisBlacklist) PutAccess(access string, createTime time.Time, expiration time.Duration) error {
	expireTime := expiration - time.Now().Sub(createTime)
	if err := p.cli.Set(p.wrapKey(access), "", expireTime).Err(); err != nil {
		return err
	}
	return nil
}

func (p *redisBlacklist) CheckAccess(access string) error {
	ttl, err := p.cli.TTL(p.wrapKey(access)).Result()
	if err != nil {
		return err
	}
	if ttl > 0 {
		return ErrExpiredAccess
	}
	return err
}

func (p *redisBlacklist) wrapKey(key string) string {
	return fmt.Sprintf("%s%s", p.prefix, key)
}
