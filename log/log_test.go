package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	l := Logger()
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
}
