package port

import (
	"matheusflix/hls-streaming-server/src/middleware"
	"net/http"
)

func SetupServer() (*http.ServeMux, error) {
	mux := http.NewServeMux()

	for endpoint, handler := range Routes {
		mux.Handle(endpoint, handler)
	}

	middleware.ServeStaticFiles(mux, "./upload", "/hls/")

	return mux, nil
}
