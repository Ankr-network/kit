package config

import (
	"github.com/caarlos0/env/v6"
	"time"
)

type Config struct {
	Verifier  Verifier
	BlackList BlackList
}

type Verifier struct {
	RSAPublicKeyPath string `env:"JWT_RSA_PUBLIC_KEY_PATH" envDefault:"/etc/ankr/secret/jwt.key.pub"`
}

type BlackList struct {
	Addr        string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password    string        `env:"REDIS_PASSWORD" envDefault:""`
	IdleTimeout time.Duration `env:"REDIS_IDLE_TIMEOUT" envDefault:"25s"`
	Prefix      string        `env:"REDIS_BLACKLIST_PREFIX" envDefault:"token:blacklist:"`
	DB          int           `env:"REDIS_BLACKLIST_DB" envDefault:"0"`
}

func LoadConfig() (*Config, error) {
	out := new(Config)
	if err := env.Parse(out); err != nil {
		return nil, err
	}
	return out, nil
}
