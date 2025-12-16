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
 // Directory containing the HLS files
 hlsDir := "./.upload"

 // Check if HLS directory exists
 if _, err := os.Stat(hlsDir); os.IsNotExist(err) {
  log.Fatalf("HLS directory %s does not exist", hlsDir)
 }

 // Serve the HLS files
 fs := http.FileServer(http.Dir(hlsDir))
 http.Handle("/hls/", http.StripPrefix("/hls/", fs))

 // Start the HTTP server with proper timeouts
 port := 8080
 server := &http.Server{
  Addr:         fmt.Sprintf(":%d", port),
  ReadTimeout:  15 * time.Second,
  WriteTimeout: 15 * time.Second,
  IdleTimeout:  60 * time.Second,
 }
 fmt.Printf("Starting HLS server on http://localhost:%d/hls/\n", port)
 log.Fatal(server.ListenAndServe())
}
