package rest

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"kit/app"
	"net/http"
)

type Server struct {
	ServeMux     *runtime.ServeMux
	Handler      http.Handler
	HttpServeMux *http.ServeMux
	Address      string
}

func NewServer(cfg *Config) *Server {
	restMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: cfg.EmitDefaults}), runtime.WithProtoErrorHandler(CustomRESTErrorHandler))

	httpServeMux := http.NewServeMux()
	httpServeMux.Handle("/", restMux)

	handler, err := RegisterCORSHandler(cfg.CORSMaxAge, httpServeMux)
	if err != nil {
		log.Fatal("register cors handler error", zap.Error(err))
	}
	return &Server{
		ServeMux:     restMux,
		Handler:      handler,
		HttpServeMux: httpServeMux,
		Address:      cfg.ListenAddress,
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
