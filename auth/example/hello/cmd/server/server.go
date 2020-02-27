package main

import (
	"context"
	"fmt"
	"github.com/Ankr-network/kit/auth"
	"github.com/Ankr-network/kit/auth/example/hello/pb"
	"github.com/Ankr-network/kit/log"
	"google.golang.org/grpc"
	"net"
)

var (
	logger = log.Logger()
)

type service struct{}

func (p *service) SayHello(ctx context.Context, req *pb.Req) (*pb.Rsp, error) {
	logger.Infof("SayHello receive: %v", req.Name)

	claim, err := auth.GetClaim(ctx)
	if err != nil {
		logger.Errorf("GetClaim error: %v", err)
	}
	logger.Infof("claim: %+v", claim)

	uid, err := auth.GetUID(ctx)
	if err != nil {
		logger.Errorf("GetUID error: %v", err)
	}
	logger.Infof("uid: %+v", uid)

	cid, err := auth.GetClientID(ctx)
	if err != nil {
		logger.Errorf("GetClientID error: %v", err)
	}
	logger.Infof("cid: %+v", cid)

	return &pb.Rsp{
		Message: fmt.Sprintf("hello %s", req.Name),
	}, nil
}

func (p *service) SayHelloInsecure(_ context.Context, req *pb.Req) (*pb.Rsp, error) {
	logger.Infof("SayHelloInsecure receive: %v", req.Name)

	return &pb.Rsp{
		Message: fmt.Sprintf("insecure hello %s", req.Name),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	bl := auth.NewRedisBlacklist(auth.NewRedisCliFromConfig())

	verifier, err := auth.NewVerifier(auth.ExcludeMethods("/pb.Hello/SayHelloInsecure"), auth.TokenBlacklist(bl))
	if err != nil {
		logger.Fatalf("newVerifier error:%v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(verifier.GRPCUnaryInterceptor()))
	pb.RegisterHelloServer(s, &service{})
	if err := s.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}
