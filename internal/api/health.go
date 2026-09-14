// Package api wires HTTP routes to their handlers.
package api

import (
	"net/http"

	"github.com/harihpdev/go-routine-example/internal/types"
)

// NewRouter builds the HTTP handler for the service.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	taskCh := make(chan types.Task, 10)
	go ProcessQueue(taskCh)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /log", logHandler)
	mux.HandleFunc("/fetch-batch", batchHandler)
	mux.HandleFunc("/fetch-channel", ChannelHandler)
	mux.HandleFunc("/enqueue", func(w http.ResponseWriter, r *http.Request) { QueueHandler(w, r, taskCh) })
	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
