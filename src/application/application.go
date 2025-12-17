package application

import (
	"matflix/hls-streaming-server/src/port"
	"net/http"
	"time"
)

func Run() error {
	handler, err := port.SetupServer()

	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
	}

	return server.ListenAndServe()
}
