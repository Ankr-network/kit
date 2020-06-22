package rest

import (
	"com.ankr.kit/util"
)

type Config struct {
	CORSMaxAge    int    `env:"CORS_MAX_AGE" envDefault:"86400"`
	ListenAddress string `env:"REST_LISTEN" envDefault:":80"`
	EmitDefaults  bool   `env:"REST_EMIT_DEFAULTS" envDefault:"true"`
}

func MustLoadConfig() *Config {
	out := new(Config)
	util.MustLoadConfig(out)
	return out
}
