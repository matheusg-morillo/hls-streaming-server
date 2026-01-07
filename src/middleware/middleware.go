package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func Chain(mws ...Middleware) Middleware {
	return func(finalHandler http.Handler) http.Handler {
		for _, mw := range mws {
			finalHandler = mw(finalHandler)
		}
		return finalHandler
	}
}

func Then(mw Middleware, handler http.Handler) http.Handler {
	return mw(handler)
}

func ThenFunc(mw Middleware, handlerFunc http.HandlerFunc) http.Handler {
	return mw(handlerFunc)
}
