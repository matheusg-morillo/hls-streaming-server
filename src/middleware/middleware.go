package middleware

import (
	"net/http"
)

func Use(before http.HandlerFunc, after http.HandlerFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if before != nil {
			before(w, r)
		}

		if r.Method != http.MethodOptions {
			next.ServeHTTP(w, r)
		}

		if after != nil {
			after(w, r)
		}
	})
}
