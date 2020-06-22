package mdb

import "com.ankr.kit/util"

type Config struct {
	URL string `env:"MONGO_URL" envDefault:"mongodb://localhost:27017"`
}

func MustLoadConfig() *Config {
	cfg := new(Config)
	util.MustLoadConfig(cfg)
	return cfg
}
