package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			log.Printf("→ %s %s", r.Method, r.URL.Path)

			next.ServeHTTP(w, r)

			duration := time.Since(start)
			log.Printf("← %s %s -> [%v]",
				r.Method, r.URL.Path, duration)
		})
	}
}
