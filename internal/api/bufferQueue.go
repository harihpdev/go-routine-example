package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/harihpdev/go-routine-example/internal/types"
)

const Max_Payload_Size = 1024 * 1024

func ProcessQueue(ch <-chan types.Task) {
	for task := range ch {
		fmt.Println("Processing job: " + task.TaskId + " Payload: " + task.Payload)
		time.Sleep(5 * time.Second)
	}
}

func QueueHandler(w http.ResponseWriter, r *http.Request, ch chan<- types.Task) {
	// Close Request Body
	defer r.Body.Close()

	var task types.Task
	payloadLimiter := io.LimitReader(r.Body, Max_Payload_Size)

	if err := json.NewDecoder(payloadLimiter).Decode(&task); err != nil {
		http.Error(w, "Cannot read payload", http.StatusBadRequest)
		return
	}

	select {
	case ch <- task:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"Accepted"}`))
	default:
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

}
