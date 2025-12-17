package application

import (
	"matflix/hls-streaming-server/src/port"
	"net/http"
)

func Run() error {
	mux, err := port.SetupServer()

	if err != nil {
		return err
	}

	return http.ListenAndServe(":8080", mux)
}
