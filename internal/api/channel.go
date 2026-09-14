package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/harihpdev/go-routine-example/internal/types"
	"github.com/harihpdev/go-routine-example/internal/util"
)

const Max_Request_Size = 1024 * 1024 // 1 MB Limit

type RequestData struct {
	Urls []string `json:"urls"`
}

func ChannelHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // Always close request body

	var wg sync.WaitGroup
	var requestData RequestData
	var result types.FetchChannelResponse

	respChan := make(chan types.FetchResult) // Unbuffered channel

	limitReader := io.LimitReader(r.Body, Max_Request_Size)
	if err := json.NewDecoder(limitReader).Decode(&requestData); err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	for _, url := range requestData.Urls {
		wg.Add(1)
		go func(u string) {
			util.CustomRequest(u, &wg, respChan)
		}(url)
	}

	// Separate closer goroutine
	go func() {
		wg.Wait()
		close(respChan)
	}()

	// Read results synchronously as workers push them into the channel
	for r := range respChan {
		if r.Err != nil {
			fmt.Printf("Error fetching %s: %v\n", r.Url, r.Err)
			continue
		}

		switch data := r.Data.(type) {
		case *types.IPResponse:
			result.Origin = data.Origin
		case *types.UUIDResponse:
			result.Uuid = data.Uuid
		case *types.UserAgentResponse:
			result.UserAgent = data.UserAgent
		default:
			fmt.Printf("Unknown data type returned from %s\n", r.Url)
		}
	}

	resBytes, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resBytes)
}
