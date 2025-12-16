package port

import "net/http"

func SetupServer(mux *http.ServeMux) error {
	for endpoint, handler := range Routes {
		mux.HandleFunc(endpoint, handler)
	}

	return nil
}
