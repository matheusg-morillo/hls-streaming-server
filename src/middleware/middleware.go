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

func Cors() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Applying CORS headers")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
		}
	})
}

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
