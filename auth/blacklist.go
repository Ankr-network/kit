package auth

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"time"
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

func NewRedisCliFromConfig() *redis.Client {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal("LoadConfig error", zap.Error(err))
	}
	return redis.NewClient(&redis.Options{
		Addr:        cfg.BlackList.Addr,
		Password:    cfg.BlackList.Password,
		DB:          cfg.BlackList.DB,
		IdleTimeout: cfg.BlackList.IdleTimeout,
	})
}

func NewRedisBlacklist(cmdable redis.Cmdable, opts ...RedisBlacklistOption) Blacklist {
	options := new(RedisBlacklistOptions)
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal("config.LoadConfig error", zap.Error(err))
	}
	options.Prefix = cfg.BlackList.Prefix

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
