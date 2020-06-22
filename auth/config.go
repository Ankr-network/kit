package auth

import (
	"com.ankr.kit/util"
	"time"
)

type VerifierConfig struct {
	RSAPublicKeyPath string `env:"JWT_RSA_PUBLIC_KEY_PATH,required"`
}

type BlackListConfig struct {
	Addr        string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password    string        `env:"REDIS_PASSWORD" envDefault:""`
	IdleTimeout time.Duration `env:"REDIS_IDLE_TIMEOUT" envDefault:"25s"`
	Prefix      string        `env:"REDIS_BLACKLIST_PREFIX" envDefault:"token:blacklist:"`
	DB          int           `env:"REDIS_BLACKLIST_DB" envDefault:"0"`
}

func MustLoadBlackListConfig() *BlackListConfig {
	out := new(BlackListConfig)
	util.MustLoadConfig(out)
	return out
}

func MustLoadVerifierConfig() *VerifierConfig {
	out := new(VerifierConfig)
	util.MustLoadConfig(out)
	return out
}
