package rest

import (
	"github.com/Ankr-network/kit/app"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	ServeMux *runtime.ServeMux
	Handler  http.Handler
	Address  string
}

func NewServerWithConfig() *Server {
	cfg := MustLoadConfig()
	restMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: cfg.EmitDefaults}), runtime.WithProtoErrorHandler(CustomRESTErrorHandler))
	handler, err := RegisterCORSHandler(restMux)
	if err != nil {
		log.Fatal("register cors handler error", zap.Error(err))
	}

	return &Server{
		ServeMux: restMux,
		Handler:  handler,
		Address:  cfg.ListenAddress,
	}
}

func (s *Server) ListenAndServed() {
	go func() {
		log.Info("start serving rest service", zap.String("address", s.Address))
		if err := http.ListenAndServe(s.Address, s.Handler); err != nil {
			app.Existing(err)
		}
	}()
}
