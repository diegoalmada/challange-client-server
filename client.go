package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type QuoResponse struct {
	Bid float64
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Println("Request timeout.")
		}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var quo QuoResponse
	err = json.Unmarshal(body, &quo)
	if err != nil {
		panic(err)
	}

	text := []byte("Dólar: " + fmt.Sprintf("%.4f", quo.Bid))
	err = os.WriteFile("cotacao.txt", text, 0644)
	if err != nil {
		panic(err)
	}
}
