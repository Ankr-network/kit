package rest

import (
	"github.com/rs/cors"
	"net/http"
)

func RegisterCORSHandler(corsMaxAge int, handler http.Handler) (http.Handler, error) {
	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return origin != ""
		},
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "DELETE", "PUT", "PATCH"},
		AllowedHeaders: []string{"Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Authorization"},
		MaxAge:         corsMaxAge,
	})

	return c.Handler(handler), nil
}
