package rpc

import (
	"github.com/Ankr-network/kit/app"
	"github.com/Ankr-network/kit/util"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthPB "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	*grpc.Server
	Address string
}

func NewServerWithConfig(interceptors ...grpc.UnaryServerInterceptor) *Server {
	return NewServer(MustLoadConfig(), interceptors...)
}

func NewServer(cfg *Config, interceptors ...grpc.UnaryServerInterceptor) *Server {
	interceptors = append(interceptors, grpcValidator.UnaryServerInterceptor())
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				interceptors...,
			),
		),
	)
	healthPB.RegisterHealthServer(s, health.NewServer())
	reflection.Register(s)
	return &Server{Server: s, Address: cfg.ListenAddress}
}

func (s *Server) MustListenAndServe() {
	lis := util.MustTcpListen(s.Address)
	go func() {
		log.Info("start serving grpc service", zap.String("address", s.Address))
		if err := s.Server.Serve(lis); err != nil {
			app.Existing(err)
		}
	}()
}
