package middleware

import (
	"log"
	"net/http"
	"os"
)

func UseStaticFiles(mux *http.ServeMux, dir string, route string) *http.ServeMux {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatalf("HLS directory %s does not exist", dir)
	}

	fs := http.FileServer(http.Dir(dir))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Cache-Control", "no-cache")

		fs.ServeHTTP(w, r)
	})

	mux.Handle("/hls/", http.StripPrefix("/hls/", Chain(CORS())(handler)))

	return mux
}
