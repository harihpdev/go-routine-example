package api

import (
	"io"
	"net/http"

	"github.com/harihpdev/go-routine-example/internal/util"
)

func logHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// TODO: handle the received log payload (write it wherever it needs to go).
	_ = body
	go util.WriteLog(body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"received"}`))
}
