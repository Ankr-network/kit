package mlog

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestNewLocal(t *testing.T) {
	logger := NewLocal()
	defer logger.Sync()

	logger.Debug("debug")
	logger.Info("info")
	logger.Error("error")
}

func TestNewServer(t *testing.T) {
	logger := NewServer()
	defer logger.Sync()

	logger.Debug("debug")
	logger.Info("info")
	logger.Error("error")
}

func TestNew(t *testing.T) {
	logger := New(MustLoadConfig())
	defer logger.Sync()

	logger.Debug("debug")
	logger.Info("info")
	logger.Error("error")
}

func TestLogger_Clone(t *testing.T) {
	a := New(MustLoadConfig())
	b := a.Clone("")

	assert.NotEqual(t, a.Logger, b.Logger)
	assert.False(t, a.level == b.level)

	a.SetLevel(zapcore.InfoLevel)
	a.Debug("debug")
	b.Debug("debug")
}

func ExampleMLog_SetLevel() {
	logger := New(MustLoadConfig())
	defer logger.Sync()

	logger.SetLevel(zapcore.InfoLevel)
	logger.Debug("info")
	// Output:
}
