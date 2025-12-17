package port

import "net/http"

func SetupServer() (*http.ServeMux, error) {
	mux := http.NewServeMux()

	for endpoint, handler := range Routes {
		mux.Handle(endpoint, handler)
	}

	return mux, nil
}
