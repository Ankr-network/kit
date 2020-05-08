package rest

import (
	"github.com/Ankr-network/kit/util"
)

type Config struct {
	CORSMaxAge    int    `env:"CORS_MAX_AGE" envDefault:"86400"`
	ListenAddress string `env:"REST_LISTEN" envDefault:":80"`
}

func MustLoadConfig() *Config {
	out := new(Config)
	util.MustLoadConfig(out)
	return out
}
