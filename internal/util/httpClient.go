package util

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/harihpdev/go-routine-example/internal/types"
)

var client = http.Client{
	Timeout: 5 * time.Second,
}

func CustomRequest(url string, mu *sync.Mutex, wg *sync.WaitGroup, ch chan<- types.FetchResult) {
	defer wg.Done()

	var res = types.FetchResult{Url: url}
	data, err := types.GetModelForURL(url)
	if err != nil {
		res.Err = err
		ch <- res
		return
	}

	resp, err := client.Get(url)
	if err != nil {
		res.Err = err
		ch <- res
		return
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		res.Err = err
		ch <- res
		return
	}
	res.Data = data
	ch <- res
}
