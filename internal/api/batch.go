package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type FetchUrls struct {
	Urls []string `json:"urls"`
}

func fetchUrl(url string, wg *sync.WaitGroup, mu *sync.Mutex, ch chan<- string) {
	defer wg.Done()

	fmt.Println("Fetching url...", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", url, err)
	}
	mu.Lock()
	// fmt.Println("Fetching url... done", url)
	ch <- "Fetched Url: " + url
	mu.Unlock()
	defer resp.Body.Close()

}

func batchHandler(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	respCh := make(chan string)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var urls FetchUrls
	err = json.Unmarshal(body, &urls)

	if err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	// fmt.Println(string(body))
	// fmt.Println(urls)

	for _, url := range urls.Urls {
		wg.Add(1)
		go (func(u string) {
			fetchUrl(u, &wg, &mu, respCh)
		})(url)
	}

	go func() {
		wg.Wait()
		close(respCh)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"received"}`))
}
