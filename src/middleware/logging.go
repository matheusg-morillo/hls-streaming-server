package middleware

import (
	"log"
	"net/http"
)

func WithInboundLogging() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Inbound request: %s %s", r.Method, r.URL.Path)
	})
}

func WithOutgoingLogging() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Outgoing response: %s %s", r.Method, r.URL.Path)
	})
}
