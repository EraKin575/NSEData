package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"server/handlers"
	"server/internal/db"
	"server/internal/processing"
	"server/models"
	"sync"
	"time"
)

func main() {
	ctx := context.Background()

	// DB connection
	connString := os.Getenv("DB_DSN")
	if connString == "" {
		connString = "postgres://optionuser:optionpass@localhost:5432/optionchain?sslmode=disable"
	}

	db, err := db.NewDB(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Shared records slice
	mu := &sync.RWMutex{}
	var records []models.ResponsePayload

	// Timezone
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}

	// HTTP handler
	http.HandleFunc("/api/data", handlers.HandlePost(&records, loc, mu))

	// Dynamic port for Railway or default local port
	port := os.Getenv("PORT")
	if port == "" {
		port = "4300"
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on :%s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Start processing option chain
	processing.ProcessingOptionChain(ctx, &records, db, loc, mu)
}
