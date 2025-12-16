package port

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func hlsHandler(w http.ResponseWriter, r *http.Request) {
	hlsDir := "./.upload"

	if _, err := os.Stat(hlsDir); os.IsNotExist(err) {
		log.Fatalf("HLS directory %s does not exist", hlsDir)
	}

	fs := http.FileServer(http.Dir(hlsDir))

	http.Handle("/hls/", http.StripPrefix("/hls/", fs))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "pong")
}

var Routes = map[string]http.HandlerFunc{
	"/hsl":         hlsHandler,
	"/checkhealth": healthHandler,
}
