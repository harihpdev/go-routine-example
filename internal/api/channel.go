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
	Urls []string
}

func channelHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup
	var mu sync.Mutex
	var requestData RequestData
	var result types.FetchChannelResponse

	respChan := make(chan types.FetchResult)

	// 1. Limit memory allocation against malicious oversized payloads
	limitReader := io.LimitReader(r.Body, Max_Request_Size)

	if err := json.NewDecoder(limitReader).Decode(&requestData); err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	for _, url := range requestData.Urls {
		wg.Add(1)
		go func(u string) {
			util.CustomRequest(u, &mu, &wg, respChan)
		}(url)
	}

	go func() {
		wg.Wait()
		close(respChan)
	}()

	for r := range respChan {

		if r.Err != nil {
			fmt.Printf("Error fetching %s: %v\n", r.Url, r.Err)
			continue
		}

		// Use a type switch to handle different formats safely
		switch data := r.Data.(type) {
		case *types.IPResponse:
			result.Origin = data.Origin
		case *types.UUIDResponse:
			result.Uuid = data.Uuid
		case *types.UserAgentResponse:
			result.UserAgent = data.UserAgent
		default:
			fmt.Printf("❓ Unknown data type returned from %s\n", r.Url)
		}
	}

	resString, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	// All success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(resString)
}
