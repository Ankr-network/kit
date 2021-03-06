package rpc

import (
	"github.com/Ankr-network/kit/util"
)

type Config struct {
	ListenAddress string `env:"GRPC_LISTEN" envDefault:":50051"`
}

func MustLoadConfig() *Config {
	out := new(Config)
	util.MustLoadConfig(out)
	return out
}
