package fetcher

import (
	"context"
	"log/slog"
	"server/internal/models"
	"strings"
	"time"
)

type FetcherService struct {
	Writer Writer
}

func (fs *FetcherService) FetchData(ctx context.Context, logger *slog.Logger) error {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		logger.Error("Failed to load location", slog.String("err", err.Error()))
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Gracefully shutdown through context cancellation")
			return nil

		case <-ticker.C:
			now := time.Now().In(loc)

			startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 15, 0, 0, loc)
			endTime := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, loc)
			resetTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, loc)
			currentDate := now.Format("02-Jan-2006")

			// Market closed today
			if isMarketHoliday(now, loc) {
				logger.Info("Market is closed today. Skipping data fetch.")
				continue
			}

			// Before market open
			if now.Before(startTime) {
				logger.Info("Market not started yet. Waiting for market to open.")
				continue
			}

			// After reset time → cleanup & wait for tomorrow
			if now.After(resetTime) {
				if err := fs.Writer.Delete(ctx); err != nil {
					logger.Error("Failed to reset stream", slog.String("err", err.Error()))
				} else {
					logger.Info("Reset stream for new trading day.")
				}
				continue
			}

			// After market close → stop until reset
			if now.After(endTime) {
				logger.Info("Market closed. Stopping data fetch until reset.")
				continue
			}

			// ---- Fetching starts here ----
			var chain models.OptionChain
			success := false
			maxRetries := 10

			for i := 0; i < maxRetries; i++ {
				var optionChainError error
				chain, optionChainError = getOptionChain()
				if optionChainError != nil {
					logger.Error("Failed to fetch option chain",
						slog.Int("attempt", i+1),
						slog.String("err", optionChainError.Error()))
					time.Sleep(2 * time.Second)
					continue
				}

				if chain.Records.TimeStamp == "" || !strings.Contains(chain.Records.TimeStamp, currentDate) {
					logger.Warn("Fetched stale data, retrying...",
						slog.Int("attempt", i+1))
					time.Sleep(2 * time.Second)
					continue
				}

				success = true
				break
			}

			if !success {
				logger.Error("All retries failed — skipping this tick")
				continue
			}

			if len(chain.Records.ExpiryDates) < 2 {
				logger.Error("Not enough expiry dates")
				continue
			}

			firstExpiry := chain.Records.ExpiryDates[0]
			secondExpiry := chain.Records.ExpiryDates[1]

			var expiryOneResult, expiryTwoResult []models.OptionData
			for _, entry := range chain.Records.Data {
				data := models.OptionData{
					StrikePrice: entry.StrikePrice,
					CE:          entry.CE,
					PE:          entry.PE,
					ExpiryDate:  entry.ExpiryDate,
				}

				switch entry.ExpiryDate {
				case firstExpiry:
					expiryOneResult = append(expiryOneResult, data)
				case secondExpiry:
					expiryTwoResult = append(expiryTwoResult, data)
				}
			}

			logger.Info("Fetched data",
				slog.Int("records_first_expiry", len(expiryOneResult)),
				slog.Int("records_second_expiry", len(expiryTwoResult)),
			)

			err = fs.Writer.Write(ctx, models.Records{
				ExpiryDates:     chain.Records.ExpiryDates,
				Data:            append(expiryOneResult, expiryTwoResult...),
				TimeStamp:       chain.Records.TimeStamp,
				UnderlyingValue: chain.Records.UnderlyingValue,
			})
			if err != nil {
				logger.Error("Failed to write to stream", slog.String("err", err.Error()))
				continue
			}

			logger.Info("Successfully wrote data to stream")
		}
	}
}
