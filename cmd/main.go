package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"server/handlers"
	"server/internal/db"
	"server/internal/processing"
	"server/models"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	mux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8090",
		Handler: mux,
	}

	go func() {
		serverErr := server.ListenAndServe()
		if serverErr != nil {
			return
		}
	}()
	connString := os.Getenv("DATABASE_PUBLIC_URL")
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

	mux.HandleFunc("/api/data", handlers.HandlePost(&records, loc, mu))

	processing.ProcessingOptionChain(ctx, &records, db, loc, mu)

	<-ctx.Done()

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutDownCtx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}


}
