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

	http.HandleFunc("/api/data", handlers.HandlePost(&records, loc, mu))

	go func() {
		log.Println("Starting server on :4300")
		if err := http.ListenAndServe(":4300", nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	processing.ProcessingOptionChain(&records, loc, mu)
}
