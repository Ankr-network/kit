package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	td := []struct {
		name string
		env  map[string]string
		cfg  *Config
		err  error
	}{
		{
			name: "empty",
			cfg: &Config{
				URL:       "amqp://guest:guest@127.0.0.1:5672",
				Exchange:  "ankr.micro",
				DLX:       "ankr.micro.dlx",
				NackDelay: 5 * time.Second,
			},
			err: nil,
		},
		{
			name: "set1",
			env: map[string]string{
				"RABBIT_URL":        "test1",
				"RABBIT_EXCHANGE":   "test2",
				"RABBIT_DLX":        "test3",
				"RABBIT_NACK_DELAY": "10s",
			},
			cfg: &Config{
				URL:       "test1",
				Exchange:  "test2",
				DLX:       "test3",
				NackDelay: 10 * time.Second,
			},
			err: nil,
		},
		{
			name: "set2",
			env: map[string]string{
				"RABBIT_URL": "test1",
			},
			cfg: &Config{
				URL:       "test1",
				Exchange:  "ankr.micro",
				DLX:       "ankr.micro.dlx",
				NackDelay: 5 * time.Second,
			},
			err: nil,
		},
		{
			name: "set3",
			env: map[string]string{
				"RABBIT_EXCHANGE": "test2",
			},
			cfg: &Config{
				URL:       "amqp://guest:guest@127.0.0.1:5672",
				Exchange:  "test2",
				DLX:       "ankr.micro.dlx",
				NackDelay: 5 * time.Second,
			},
			err: nil,
		},
	}

	for _, d := range td {
		t.Run(d.name, func(t *testing.T) {
			for k, v := range d.env {
				err := os.Setenv(k, v)
				require.NoError(t, err)
			}

			out, err := LoadConfig()
			assert.Equal(t, d.cfg, out)
			assert.Equal(t, d.err, err)

			for k, _ := range d.env {
				err := os.Unsetenv(k)
				require.NoError(t, err)
			}
		})
	}
}
