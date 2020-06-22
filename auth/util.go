package auth

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"google.golang.org/grpc/metadata"
	"strings"
)

var (
	ErrInvalidContext = errors.New("context without authentication")
)

func NewRedisCli(cfg *BlackListConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:        cfg.Addr,
		Password:    cfg.Password,
		DB:          cfg.DB,
		IdleTimeout: cfg.IdleTimeout,
	})
}

func GetUID(ctx context.Context) (string, error) {
	claim, err := GetClaim(ctx)
	if err != nil {
		return "", err
	}

	val, ok := claim["sub"]
	if !ok {
		return "", ErrInvalidContext
	}
	uid, ok := val.(string)
	if !ok {
		return "", ErrInvalidContext
	}
	return uid, nil
}

func GetClientID(ctx context.Context) (string, error) {
	claim, err := GetClaim(ctx)
	if err != nil {
		return "", err
	}

	val, ok := claim["aud"]
	if !ok {
		return "", ErrInvalidContext
	}
	cid, ok := val.(string)
	if !ok {
		return "", ErrInvalidContext
	}
	return cid, nil
}

func GetClaim(ctx context.Context) (jwt.MapClaims, error) {
	claim, ok := ctx.Value("claim").(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidContext
	}
	return claim, nil
}

func ExtractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrMissingMetadata
	}

	array := md["authorization"]
	if len(array) < 1 {
		return "", ErrEmptyAuthorization
	}

	return strings.TrimPrefix(array[0], "Bearer "), nil
}
