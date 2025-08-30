package main

import (
	"context"
	"log/slog"
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

func initLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	logger := initLogger()
	defer stop()

	mux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8090",
		Handler: mux,
	}

	go func() {
		logger.Info("Starting server on:", server.Addr)
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
		logger.Error("Failed to connect to DB:", slog.String("error", err.Error()))
	}

	// Shared records slice
	mu := &sync.RWMutex{}
	var records []models.ResponsePayload

	// Timezone
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		logger.Error("Failed to load timezone:", slog.String("error", err.Error()))
	}

	mux.HandleFunc("/api/data", handlers.HandlePost(&records, loc, mu, logger))

	processing.ProcessingOptionChain(ctx, &records, db, loc, mu, logger)

	<-ctx.Done()

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutDownCtx); err != nil {
		logger.Error("Error shutting down server:", slog.String("error", err.Error()))
	}

}
