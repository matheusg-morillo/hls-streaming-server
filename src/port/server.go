package port

import (
	"errors"
	"matflix/hls-streaming-server/src/middleware"
	"net/http"
	"os"
)

func SetupServer() (*http.ServeMux, error) {
	mux := http.NewServeMux()
	hlsDir := os.Getenv("HLS_DIR")

	if hlsDir == "" {
		return nil, errors.New("HLS_DIR environment variable is not set")
	}

	for endpoint, handler := range Routes {
		mux.Handle(endpoint, handler)
	}

	middleware.UseStaticFiles(mux, hlsDir, "/hls/")

	return mux, nil
}
