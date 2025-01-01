package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type UpdatePostPayload struct {
	Title   string `json:"title" validate:"omitempty,max=200"`
	Content string `json:"content" validate:"omitempty,max=1000"`
}

func updatePost(postID int, p UpdatePostPayload, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("http://localhost:8080/v1/posts/%d", postID)

	b, _ := json.Marshal(p)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("Error creating request", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request", err)
		return
	}

	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	postID := 4

	go updatePost(postID, UpdatePostPayload{Title: "Updated Title"}, &wg)
	go updatePost(postID, UpdatePostPayload{Content: "Updated Content"}, &wg)

	wg.Wait()
}
