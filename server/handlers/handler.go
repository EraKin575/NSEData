package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/models"
	"sync"
	"time"
)

func HandlePost(data *models.Records, loc *time.Location, endTime time.Time, mu *sync.RWMutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher := w.(http.Flusher)

		if r.Method == http.MethodOptions {
			// For preflight requests, return a 200 and no body.
			w.WriteHeader(http.StatusOK)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		for time.Now().In(loc).Before(endTime) {
			mu.RLock()
			jsonRecords, err := json.Marshal(data)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error marshalling data: %v", err), http.StatusInternalServerError)
				return
			}
			mu.RUnlock()

			fmt.Fprintf(w, "data: %s\n\n", jsonRecords)
			flusher.Flush()
			time.Sleep(1 * time.Second)
		}

	}
	// Return the header with the values copied into a shared backing array.
}
