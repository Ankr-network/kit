package trace

import (
	"github.com/Ankr-network/kit/util"
	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"strings"
)

type Config struct {
	ServiceName                         string        `env:"JAEGER_SERVICE_NAME,required"`
	RPCMetrics                          bool          `env:"JAEGER_RPC_METRICS"`
	Disabled                            bool          `env:"JAEGER_DISABLED" envDefault:"false"`
	Tags                                []string      `env:"JAEGER_TAGS" envSeparator:","` // key1=value1,key2=value2
	SamplerType                         string        `env:"JAEGER_SAMPLER_TYPE" envDefault:"const"`
	SamplerParam                        float64       `env:"JAEGER_SAMPLER_PARAM" encDefault:"1"`
	ReporterLogSpans                    bool          `env:"JAEGER_REPORTER_LOG_SPANS" envDefault:"true"`
	LocalAgentHostPort                  string        `env:"JAEGER_AGENT" envDefault:"127.0.0.1"`
	//SamplerManagerHostPort              string        `env:"JAEGER_SAMPLER_MANAGER_HOST_PORT"`
	//SamplingEndpoint                    string        `env:"JAEGER_ENDPOINT"`
	//SamplerMaxOperations                int           `env:"JAEGER_SAMPLER_MAX_OPERATIONS"`
	//SamplerRefreshInterval              time.Duration `env:"JAEGER_SAMPLER_REFRESH_INTERVAL"`
	//ReporterMaxQueueSize                int           `env:"JAEGER_REPORTER_MAX_QUEUE_SIZE"`
	//ReporterFlushInterval               time.Duration `env:"JAEGER_REPORTER_FLUSH_INTERVAL"`
	//ReporterAttemptReconnectingDisabled bool          `env:"JAEGER_REPORTER_ATTEMPT_RECONNECTING_DISABLED"`
	//ReporterAttemptReconnectInterval    time.Duration `env:"JAEGER_REPORTER_ATTEMPT_RECONNECT_INTERVAL"`
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
			Type:                    c.SamplerType,
			Param:                   c.SamplerParam,
			//SamplingServerURL:       c.SamplingEndpoint,
			//SamplingRefreshInterval: c.ReporterFlushInterval,
			//MaxOperations:           c.SamplerMaxOperations,
		},
		Reporter: &jaegercfg.ReporterConfig{
			//QueueSize:                  c.ReporterMaxQueueSize,
			//BufferFlushInterval:        c.ReporterFlushInterval,
			LogSpans:                   c.ReporterLogSpans,
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
