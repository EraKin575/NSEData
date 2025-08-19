package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/models"
	"sync"
	"time"
)

// HandlePost streams `data` as SSE until `endTime`.
// Uses a mutex for safe concurrent access to `data`.
func HandlePost(data *[]models.Records, loc *time.Location, mu *sync.RWMutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS and preflight request
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			// Preflight request (CORS)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set headers for Server-Sent Events (SSE)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// Stream data every second until endTime
		for {
			now := time.Now().In(loc)
			endTime := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc)

			if now.After(endTime) {
				log.Println("SSE stream stopped: reached 8 PM")
				break
			}

			// Safely read data with mutex
			mu.RLock()
			jsonRecords, err := json.Marshal(data)
			mu.RUnlock()

			if err != nil {
				http.Error(w, fmt.Sprintf("Error marshalling data: %v", err), http.StatusInternalServerError)
				return
			}

			// Write SSE-formatted data
			fmt.Fprintf(w, "data: %s\n\n", jsonRecords)
			flusher.Flush()

			// Sleep for 3 minutes before next push
			time.Sleep(3 * time.Minute)
		}
	}
}
