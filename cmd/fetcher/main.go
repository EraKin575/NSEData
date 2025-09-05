package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"server/internal/fetcher"
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
	defer stop()

	logger := initLogger()

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}	

	redisClient := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	writer := fetcher.NewStreamWriter(redisClient, "nifty50:option_chain")
	fetcherService := &fetcher.FetcherService{
		Writer: writer,
	}

	if err := fetcherService.FetchData(ctx, logger); err != nil {
		logger.Error("Failed to fetch data", slog.String("err", err.Error()))
	}

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	// Graceful shutdown context (if you had cleanup tasks)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Example cleanup: close Redis
	if err := redisClient.Close(); err != nil {
		logger.Error("Failed to close Redis", slog.String("err", err.Error()))
	}

	<-shutdownCtx.Done()
}
