package application

import (
	"matheusflix/hls-streaming-server/src/port"
	"net/http"
)

func Run() error {
	mux := http.NewServeMux()

	err := port.SetupServer(mux)

	if err != nil {
		return err
	}

	http.ListenAndServe(":8080", mux)

	return nil
}
