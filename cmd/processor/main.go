package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"server/handlers"
	"server/internal/db"
	"server/internal/models"
	"server/internal/processing"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func initLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	logger := initLogger()
	defer stop()

	records := &[]models.ResponsePayload{}

	mu := &sync.RWMutex{}
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		logger.Error("Failed to load location", slog.String("error", err.Error()))
		return
	}

	mux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8090",
		Handler: mux,
	}

	go func() {
		logger.Info("Starting server on :8090")
		serverErr := server.ListenAndServe()
		if serverErr != nil {
			return
		}
	}()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://optionuser:optionpass@localhost:5432/optionchain?sslmode=disable"
	}

	db, err := db.NewDB(ctx, connString)
	if err != nil {
		logger.Error("Failed to connect to DB:", slog.String("error", err.Error()))
	}

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	reader := processing.NewStreamReader(redisClient, "nifty50:option_chain")
	processingService := &processing.ProcessingService{
		Reader:   reader,
		DBWriter: db,
	}

	if err := processingService.ProcessingOptionChain(ctx, db, mu, logger, records); err != nil {
		logger.Error("Failed to process data", slog.String("err", err.Error()))
	}
	mux.HandleFunc("/api/data", handlers.HandlePost(records, loc, mu, logger))

	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Close(); err != nil {
		logger.Error("Failed to close Redis", slog.String("err", err.Error()))
	}

	<-shutdownCtx.Done()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown server", slog.String("err", err.Error()))
	}
	logger.Info("Server shutdown complete")

}
