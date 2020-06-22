package rdb

import (
	"kit.self/util"
	"time"
)

type Config struct {
	DSN             string        `env:"MYSQL_DSN,required"`
	ConnMaxLifetime time.Duration `env:"MYSQL_CONN_MAX_TIME" envDefault:"30m"`
	MaxIdleConns    int           `env:"MYSQL_CONN_MAX_IDLE" envDefault:"10"`
	SetMaxOpenConns int           `env:"MYSQL_CONN_MAX_OPEN" envDefault:"100"`
}

func MustLoadConfig() *Config {
	out := new(Config)
	util.MustLoadConfig(out)
	return out
}
