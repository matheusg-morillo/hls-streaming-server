package port

import (
	"encoding/json"
	"log"
	"matheusflix/hls-streaming-server/src/wire/out"
	"net/http"
	"os"
	"time"
)

func hlsHandler(hlsDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(hlsDir); os.IsNotExist(err) {
			log.Fatalf("HLS directory %s does not exist", hlsDir)
		}

		fs := http.FileServer(http.Dir(hlsDir))

		http.Handle("/hls/", http.StripPrefix("/hls/", fs))
	})
}

func healthHandler(timeFormat string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkHealthOut := out.CheckHealth{
			Status: "healthy",
			Time:   time.Now().Format(timeFormat),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(checkHealthOut); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

var Routes = map[string]http.Handler{
	"/hsl":         hlsHandler("./upload"),
	"/checkhealth": healthHandler(time.RFC1123),
}
