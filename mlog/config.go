package mlog

import (
	"kit.self/util"
)

type Config struct {
	Mode Mode `env:"LOG_MODE" envDefault:"local"`
}

func MustLoadConfig() *Config {
	out := new(Config)
	util.MustLoadConfig(out)
	return out
}
