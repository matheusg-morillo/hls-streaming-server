// Package main provides a simple HTTP server for streaming HLS video content.
package main

import (
 "fmt"
 "log"
 "net/http"
 "os"
 "time"
)

func main() {
	hlsDir := "./.upload"

	if _, err := os.Stat(hlsDir); os.IsNotExist(err) {
		log.Fatalf("HLS directory %s does not exist", hlsDir)
	}

	fs := http.FileServer(http.Dir(hlsDir))

	http.Handle("/hls/", http.StripPrefix("/hls/", fs))
	http.HandleFunc("/ping/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "pong")
	})

	port := 8080
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Starting HLS server on http://localhost:%d/\n", port)
	log.Fatal(server.ListenAndServe())
}
