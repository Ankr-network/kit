package log

import "github.com/caarlos0/env/v6"

type Config struct {
	Level        string `env:"LOG_LEVEL" envDefault:"INFO"`
	ReportCaller bool   `env:"LOG_REPORT_CALLER" envDefault:"true"`
}

func LoadConfig() (*Config, error) {
	out := new(Config)
	if err := env.Parse(out); err != nil {
		return nil, err
	}
	return out, nil
}
