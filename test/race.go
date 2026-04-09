package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
)

func main() {
	url := "http://127.0.0.1:4000/v1/movies/4"
	jsonBody := []byte(`{"runtime": "105 mins"}`)

	var wg sync.WaitGroup
	// Bắn 20 yêu cầu cùng lúc bằng Goroutines
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(jsonBody))
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Request %d: Error %v\n", id, err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("Request %d: Status %s\n", id, resp.Status)
		}(i)
	}
	wg.Wait()
}
