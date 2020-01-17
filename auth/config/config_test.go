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
			name: "full set",
			env: map[string]string{
				"JWT_RSA_PUBLIC_KEY_PATH": "test1",
				"REDIS_ADDR":              "test2",
				"REDIS_PASSWORD":          "test3",
				"REDIS_IDLE_TIMEOUT":      "30s",
				"REDIS_BLACKLIST_PREFIX":  "test4",
				"REDIS_BLACKLIST_DB":      "1",
			},
			cfg: &Config{
				Verifier: Verifier{RSAPublicKeyPath: "test1"},
				BlackList: BlackList{
					Addr:        "test2",
					Password:    "test3",
					IdleTimeout: 30 * time.Second,
					Prefix:      "test4",
					DB:          1,
				},
			},
			err: nil,
		},
		{
			name: "empty",
			env:  map[string]string{},
			cfg: &Config{
				Verifier: Verifier{RSAPublicKeyPath: "/etc/ankr/secret/jwt.key.pub"},
				BlackList: BlackList{
					Addr:        "localhost:6379",
					Password:    "",
					IdleTimeout: 25 * time.Second,
					Prefix:      "token:blacklist:",
					DB:          0,
				},
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

			for k := range d.env {
				err := os.Unsetenv(k)
				require.NoError(t, err)
			}
		})
	}
}
