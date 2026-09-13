// Package api wires HTTP routes to their handlers.
package api

import "net/http"

// NewRouter builds the HTTP handler for the service.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /log", logHandler)
	mux.HandleFunc("/fetch-batch", batchHandler)
	mux.HandleFunc("/fetch-channel", channelHandler)
	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
