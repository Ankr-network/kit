package mlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
)

type Mode string

const (
	ModeLocal  Mode = "local"
	ModeServer      = "server"
)

type MLog struct {
	cfg  zap.Config
	opts []zap.Option
	*zap.Logger
	level zap.AtomicLevel
}

func (m *MLog) SetLevel(level zapcore.Level) {
	m.level.SetLevel(level)
}

func (m *MLog) Clone(name string) *MLog {
	out := &MLog{
		cfg:   m.cfg,
		opts:  m.opts,
		level: zap.NewAtomicLevelAt(m.level.Level()),
	}
	out.cfg.Level = out.level
	nl, err := out.cfg.Build(out.opts...)
	if err != nil {
		log.Panicf("cfg.Build error: %v", err)
	}
	out.Logger = nl.Named(name)
	return out
}

func (m *MLog) Sync() {
	_ = m.Logger.Sync()
}

func (m *MLog) Ignore(err error) {
	if err != nil {
		m.Error("ignore error", zap.Error(err))
	}
}

func (m *MLog) Close(c io.Closer) {
	m.Ignore(c.Close())
}

// NewServer return a logger for server online mode
func NewServer(opts ...zap.Option) *MLog {
	cfg := NewServerConfig()
	logger, err := cfg.Build(opts...)
	if err != nil {
		log.Panicf("cfg.Build error: %v", err)
	}
	return &MLog{
		cfg:    cfg,
		opts:   opts,
		Logger: logger,
		level:  cfg.Level,
	}
}

// NewLocal return a logger for local development mode
func NewLocal(opts ...zap.Option) *MLog {
	cfg := NewLocalConfig()
	logger, err := cfg.Build(opts...)
	if err != nil {
		log.Panicf("cfg.Build error: %v", err)
	}
	return &MLog{
		cfg:    cfg,
		opts:   opts,
		Logger: logger,
		level:  cfg.Level,
	}
}

// New return a logger with env configuration
func New(cfg *Config) *MLog {
	switch cfg.Mode {
	case ModeLocal:
		return NewLocal()
	case ModeServer:
		return NewServer()
	default:
		panic("error: invalid Mode")
	}
}

func NewServerConfig() zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg
}

func NewLocalConfig() zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg
}
