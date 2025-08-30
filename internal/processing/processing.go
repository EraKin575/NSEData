package processing

import (
	"context"
	"log/slog"
	"math"
	"server/internal/api"
	"server/internal/db"
	"server/models"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

func ProcessingOptionChain(ctx context.Context, records *[]models.ResponsePayload, db *db.DB, loc *time.Location, mu *sync.RWMutex, logger *slog.Logger) {
	c := cron.New(
		cron.WithLocation(loc),
		cron.WithSeconds(), // allow seconds field in cron expression
	)

	var lastTimeStampRecorded string

	// --- 1. Fetch job every 3 mins between 09:15 and 15:39 ---
	_, err := c.AddFunc("0 15-39/3 9-15 * * MON-FRI", func() {
		now := time.Now().In(loc)
		resetTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 55, 0, 0, loc)

		// Fetch fresh records
		newRecords, err := api.FetchData(logger)
		if err != nil {
			logger.Error("Failed to fetch records:", slog.String("error", err.Error()))
			return
		}
		if newRecords.TimeStamp == "" {
			logger.Error("Empty timestamp received")
			return
		}

		// Ensure timestamp is valid and unique
		timestampParts := strings.Split(newRecords.TimeStamp, " ")
		if len(timestampParts) == 0 {
			logger.Error("Invalid timestamp format:", slog.String("timestamp", newRecords.TimeStamp))
			return
		}
		datePart := timestampParts[0]
		expectedDate := now.Format("02-Jan-2006")

		if datePart == expectedDate && lastTimeStampRecorded != newRecords.TimeStamp {
			mu.Lock()
			*records = append(*records, extractResponsePayload(newRecords)...)
			mu.Unlock()
			lastTimeStampRecorded = newRecords.TimeStamp
			logger.InfoContext(ctx, "Record added successfully", slog.String("timestamp", newRecords.TimeStamp))
		} else {
			logger.InfoContext(ctx, "Duplicate or invalid timestamp, record not added", slog.String("timestamp", newRecords.TimeStamp))

		}

		if now.After(resetTime) {
			mu.Lock()
			*records = []models.ResponsePayload{}
			mu.Unlock()
			lastTimeStampRecorded = "" // reset timestamp
			logger.InfoContext(ctx, "Records reset for the next day")
		}
	})

	if err != nil {
		logger.Error("Failed to add fetch job:", slog.String("error", err.Error()))
	}

	_, err = c.AddFunc("0 40 15 * * MON-FRI", func() {
		now := time.Now().In(loc)
		mu.RLock()
		defer mu.RUnlock()

		if len(*records) == 0 {
			logger.InfoContext(ctx, "No records to write at close")
			return
		}

		if err := db.WriteToDB(ctx, *records); err != nil {
			logger.Error("Error writing to DB:", slog.String("error", err.Error()))
		} else {
			logger.InfoContext(ctx, "Final data written to DB", slog.String("time", now.String()))
		}
	})

	if err != nil {
		logger.Error("Failed to add final DB write job:", slog.String("error", err.Error()))
	}

	c.Start()
	defer c.Stop()

	<-ctx.Done()
	logger.Debug("Processing stopped")
}

func extractResponsePayload(records models.Records) []models.ResponsePayload {
	var response []models.ResponsePayload
	for _, record := range records.Data {
		ceOI, ceChOI, ceVol, ceIV, ceLTP := 0.0, 0.0, 0, 0.0, 0.0
		peOI, peChOI, peVol, peIV, peLTP := 0.0, 0.0, 0, 0.0, 0.0

		if record.CE != nil {
			ceOI = record.CE.OpenInterest
			ceChOI = record.CE.ChangeInOpenInterest
			ceVol = record.CE.TotalTradedVolume
			ceIV = record.CE.ImpliedVolatility
			ceLTP = record.CE.LastPrice
		}
		if record.PE != nil {
			peOI = record.PE.OpenInterest
			peChOI = record.PE.ChangeInOpenInterest
			peVol = record.PE.TotalTradedVolume
			peIV = record.PE.ImpliedVolatility
			peLTP = record.PE.LastPrice
		}

		pcr := calculatePCR(peOI, ceOI)
		intradayPCR := calculatePCR(peChOI, ceChOI)
		ceChOIPercentage := calculatePercentage(ceChOI, ceOI)
		peChOIPercentage := calculatePercentage(peChOI, peOI)

		timeStamp, err := time.Parse("02-Jan-2006 15:04:05", records.TimeStamp)
		if err != nil {
			timeStamp = time.Time{}
		}
		expiryDate, err := time.Parse("02-Jan-2006", record.ExpiryDate)
		if err != nil {
			expiryDate = time.Time{}
		}

		response = append(response, models.ResponsePayload{
			Timestamp:                        timeStamp,
			ExpiryDate:                       expiryDate,
			StrikePrice:                      record.StrikePrice,
			UnderlyingValue:                  records.UnderlyingValue,
			CEOpenInterest:                   ceOI,
			CEChangeInOpenInterest:           ceChOI,
			CEChangeInOpenInterestPercentage: ceChOIPercentage,
			CETotalTradedVolume:              ceVol,
			CEImpliedVolatility:              ceIV,
			CELastPrice:                      ceLTP,
			PEOpenInterest:                   peOI,
			PEChangeInOpenInterest:           peChOI,
			PEChangeInOpenInterestPercentage: peChOIPercentage,
			PETotalTradedVolume:              peVol,
			PEImpliedVolatility:              peIV,
			PELastPrice:                      peLTP,
			PCR:                              pcr,
			IntraDayPCR:                      intradayPCR,
		})
	}
	return response
}

func calculatePCR(num, denom float64) float64 {
	if denom == 0 || !isFinite(num/denom) {
		return -1
	}
	return num / denom
}

func calculatePercentage(changeOI, baseOI float64) float64 {
	if baseOI == 0 {
		return 0
	}
	return (changeOI / baseOI) * 100
}

func isFinite(value float64) bool {
	return !math.IsInf(value, 0) && !math.IsNaN(value)
}
