package reporter

import 	"github.com/caarlos0/env/v6"


type Config struct {
	MixpanelToken string `env:"MIXPANEL_TOKEN,required"`
}

func LoadConfig() (*Config, error) {
	out := new(Config)
	if err := env.Parse(out); err != nil {
		return nil, err
	}
	return out, nil
}
