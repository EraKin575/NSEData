package main

import (
	"log"
	"net/http"
	"server/internal/api"
	"server/handlers"
	"server/models"
	"sync"
	"time"
)

func main() {
	mu := &sync.RWMutex{}
	var records models.Records
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}

	now := time.Now().In(loc)
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 21, 59, 0, 0, loc)

	http.HandleFunc("/api/data", handlers.HandlePost(&records, loc, endTime, mu))

	go func() {
		log.Println("Starting server on :4300")
		if err := http.ListenAndServe(":4300", nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	startTime := time.Date(now.Year(), now.Month(), now.Day(), 00, 01, 0, 0, loc)

	if now.Before(startTime) {
		wait := time.Until(startTime)
		log.Printf("Waiting until market open at %s (%v)...", startTime.Format("15:04:05"), wait)
		time.Sleep(wait)
	}

	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		// Check time in IST consistently
		if time.Now().In(loc).After(endTime) {
			log.Println("Reached 16:15 IST. Stopping fetch loop.")
			break
		}

		wantedRecords, err := api.FetchData()
		if err != nil {
			log.Printf("Error fetching data: %v", err)
			continue
		}
		if len(wantedRecords.Data) == 0 {
			log.Println("No data fetched, retrying...")
			continue
		}

		mu.Lock()
		
        // rest fields are static, so we can append directly
		records = models.Records{
			TimeStamp:      wantedRecords.TimeStamp,
			UnderlyingValue: wantedRecords.UnderlyingValue,
			Data:           append(records.Data, wantedRecords.Data...),
			ExpiryDates:    wantedRecords.ExpiryDates,
		}

		mu.Unlock()

		log.Printf("Fetched %d option records", len(wantedRecords.Data))
		<-ticker.C
	}

	log.Println("Market closed. Server continues running...")
	select {}
}
