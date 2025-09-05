package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"server/internal/models"
	"sync"
	"time"
)

// HandlePost streams `data` as SSE until `endTime`.
// Uses a mutex for safe concurrent access to `data`.
func HandlePost(data *[]models.ResponsePayload, loc *time.Location, mu *sync.RWMutex, logger *slog.Logger) http.HandlerFunc {
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
			logger.Error("Streaming unsupported!")
			return
		}

		for {
			now := time.Now().In(loc)
			endTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, loc)

			if now.After(endTime) {
				log.Println("SSE stream paused")
				break
			}

			mu.RLock()
			jsonRecords, err := json.Marshal(data)
			mu.RUnlock()

			if err != nil {
				logger.Error("Error marshalling records:", slog.String("error", err.Error())	)
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", jsonRecords)
			flusher.Flush()

			time.Sleep(3 * time.Minute)
		}
	}
}
