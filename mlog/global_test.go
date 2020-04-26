package mlog

import (
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestLogger(t *testing.T) {
	root := Logger("")
	root.Info("info")

	sub := Logger("sub")
	sub.Info("sub")
}

func TestMLog_SetLevel(t *testing.T) {
	root := Logger("")
	sub := Logger("sub")

	root.Debug("debug 1")
	sub.Debug("debug 1")

	root.SetLevel(zapcore.InfoLevel)
	root.Debug("debug 2")
	sub.Debug("debug 2")

	sub.SetLevel(zapcore.InfoLevel)
	root.SetLevel(zapcore.DebugLevel)
	root.Debug("debug 3")
	sub.Debug("debug 3")
}
