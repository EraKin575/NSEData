package main

import (
	"log"
	"net/http"
	"server/handlers"
	"server/internal/processing"
	"server/models"
	"sync"
	"time"
)

func main() {
	mu := &sync.RWMutex{}
	var records []models.Records
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

	processing.ProcessingOptionChain(&records, startTime, endTime, loc, mu)
}
