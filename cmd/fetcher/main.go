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

func initRedis(logger *slog.Logger) *goredis.Client {
	redisURL := os.Getenv("REDIS_URL")
	var client *goredis.Client

	switch {
	case redisURL == "":
		// Local fallback
		client = goredis.NewClient(&goredis.Options{
			Addr: "localhost:6379",
		})

	case len(redisURL) >= 8 && redisURL[:8] == "redis://":
		// Railway style URL
		opt, err := goredis.ParseURL(redisURL)
		if err != nil {
			logger.Error("Failed to parse Redis URL", slog.String("err", err.Error()))
			os.Exit(1)
		}
		client = goredis.NewClient(opt)

	default:
		// Assume host:port only
		client = goredis.NewClient(&goredis.Options{
			Addr: redisURL,
		})
	}

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Error("Failed to connect to Redis", slog.String("err", err.Error()))
		os.Exit(1)
	}

	logger.Info("Connected to Redis")
	return client
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger := initLogger()

	redisClient := initRedis(logger)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis", slog.String("err", err.Error()))
		}
	}()

	writer := fetcher.NewStreamWriter(redisClient, "nifty50:option_chain")

	browser := fetcher.NewBrowser()
	defer browser.Close()

	fetcherService := &fetcher.FetcherService{
		Writer:  writer,
		Browser: browser,
	}

	if err := fetcherService.FetchData(ctx, logger); err != nil {
		logger.Error("Failed to fetch data", slog.String("err", err.Error()))
	}

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	// Graceful shutdown context
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-shutdownCtx.Done()
}
