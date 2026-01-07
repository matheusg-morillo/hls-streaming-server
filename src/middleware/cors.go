package middleware

import (
	"log"
	"net/http"
	"strings"
)

type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

func isPreflightRequest(r *http.Request) bool {
	return r.Method == http.MethodOptions
}

func CORSWithConfig(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("→ Applying CORS headers")

			w.Header().Set("Access-Control-Allow-Origin", strings.Join(config.AllowOrigins, ","))
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ","))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ","))
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")

			if isPreflightRequest(r) {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)

			log.Println("← Applied CORS headers")
		})
	}
}

func CORS() Middleware {
	return CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"*"},
	})
}
