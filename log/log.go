package log

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

var (
	defaultLog = logrus.New()
)

func init() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("LoadConfig error: %v", err)
	}

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		log.Fatalf("logrus.ParseLevel error: %v", err)
	}

	defaultLog.SetLevel(level)
	defaultLog.SetReportCaller(cfg.ReportCaller)
}

func Logger() *logrus.Logger {
	return defaultLog
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SetLevel(_ context.Context, req *SetLevelRequest) (*SetLevelResponse, error) {
	defaultLog.Infof("Handle SetLevel rpc %v", req)
	lv, err := logrus.ParseLevel(req.Level.String())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defaultLog.SetLevel(lv)
	return &SetLevelResponse{
		Level: req.Level,
	}, nil
}
