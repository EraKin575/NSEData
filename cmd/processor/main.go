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

func initRedis(ctx context.Context, logger *slog.Logger) *goredis.Client {
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

	// Connectivity test
	if err := client.Ping(ctx).Err(); err != nil {
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

	records := &[]models.ResponsePayload{}
	mu := &sync.RWMutex{}

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		logger.Error("Failed to load location", slog.String("error", err.Error()))
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		logger.Info("Starting server", slog.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", slog.String("err", err.Error()))
		}
	}()

	// --- Database setup ---
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://optionuser:optionpass@localhost:5432/optionchain?sslmode=disable"
	}
	db, err := db.NewDB(ctx, connString)
	if err != nil {
		logger.Error("Failed to connect to DB", slog.String("error", err.Error()))
	}

	// --- Redis setup ---
	redisClient := initRedis(ctx, logger)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis", slog.String("err", err.Error()))
		}
	}()

	// --- Processing service ---
	reader := processing.NewStreamReader(redisClient, "nifty50:option_chain")
	processingService := &processing.ProcessingService{
		Reader:   reader,
		DBWriter: db,
	}

	mux.HandleFunc("/api/data", handlers.HandlePost(records, loc, mu, logger))

	if err := processingService.ProcessingOptionChain(ctx, db, mu, logger, records); err != nil {
		logger.Error("Failed to process data", slog.String("err", err.Error()))
	}

	// --- Shutdown handling ---
	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown server", slog.String("err", err.Error()))
	}
	logger.Info("Server shutdown complete")
}
