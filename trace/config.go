package trace

import (
	"fmt"
	"github.com/Ankr-network/kit/util"
	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"strings"
)

type Config struct {
	ServiceName      string   `env:"JAEGER_SERVICE_NAME,required"`
	RPCMetrics       bool     `env:"JAEGER_RPC_METRICS"`
	Disabled         bool     `env:"JAEGER_DISABLED" envDefault:"false"`
	Tags             []string `env:"JAEGER_TAGS" envSeparator:","` // key1=value1,key2=value2
	SamplerType      string   `env:"JAEGER_SAMPLER_TYPE" envDefault:"const"`
	SamplerParam     float64  `env:"JAEGER_SAMPLER_PARAM" envDefault:"1"`
	ReporterLogSpans bool     `env:"JAEGER_REPORTER_LOG_SPANS" envDefault:"true"`
	LocalAgentHost   string   `env:"JAEGER_AGENT_HOST" envDefault:"127.0.0.1"`
	LocalAgentPort   string   `env:"JAEGER_AGENT_PORT" envDefault:"6831"`
}

func MustLoadConfig() *Config {
	out := new(Config)
	util.MustLoadConfig(out)
	return out
}

func (c *Config) ToTraceConfiguration() *jaegercfg.Configuration {
	config := &jaegercfg.Configuration{
		ServiceName: c.ServiceName,
		Disabled:    c.Disabled,
		RPCMetrics:  c.RPCMetrics,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  c.SamplerType,
			Param: c.SamplerParam,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           c.ReporterLogSpans,
			LocalAgentHostPort: fmt.Sprintf("%s:%s", c.LocalAgentHost, c.LocalAgentPort),
		},
	}

	for _, v := range c.Tags {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) == 2 {
			config.Tags = append(config.Tags, opentracing.Tag{
				Key:   kv[0],
				Value: kv[1],
			})
		}
	}
	return config
}
