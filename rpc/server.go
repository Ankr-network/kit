package rpc

import (
	"github.com/Ankr-network/kit/app"
	"github.com/Ankr-network/kit/auth"
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

func NewBlackListServerWithConfig(bl auth.Blacklist, additionalExcludeMethods ...string) *Server {
	excludeMethods := []string{
		"/.+Internal.+/.+",
		"/grpc.health.v1.Health/Check",
	}
	excludeMethods = append(excludeMethods, additionalExcludeMethods...)
	verifier, err := auth.NewVerifier(auth.ExcludeMethods(excludeMethods...), auth.TokenBlacklist(bl))
	if err != nil {
		log.Fatal("NewVerifier error", zap.Error(err))
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				verifier.GRPCUnaryInterceptor(),
				grpcValidator.UnaryServerInterceptor(),
			),
		),
	)
	healthPB.RegisterHealthServer(s, health.NewServer())
	reflection.Register(s)
	return &Server{Server: s, Address: MustLoadConfig().ListenAddress}
}

func NewServerWithConfig(additionalExcludeMethods ...string) *Server {
	return NewBlackListServerWithConfig(nil, additionalExcludeMethods...)
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
