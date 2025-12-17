package port

import (
	"encoding/json"
	"matflix/hls-streaming-server/src/adapter"
	"matflix/hls-streaming-server/src/controller"
	"matflix/hls-streaming-server/src/middleware"
	"net/http"
)

func healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := controller.Health()

		if err := json.NewEncoder(w).Encode(adapter.HealthToJSON(response)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

var Routes = map[string]http.Handler{
	"/health": middleware.Use(middleware.WithInboundLogging(), middleware.WithOutgoingLogging(), healthHandler()),
}
