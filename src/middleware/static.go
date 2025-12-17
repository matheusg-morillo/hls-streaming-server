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

	mux.Handle("/hls/", http.StripPrefix("/hls/", fs))

	return mux
}
