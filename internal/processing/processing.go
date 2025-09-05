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
)

func ProcessingOptionChain(ctx context.Context, records *[]models.ResponsePayload, db *db.DB, loc *time.Location, mu *sync.RWMutex, logger *slog.Logger) {
	var lastTimeStampRecorded string
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()
	isWrittenToDB := false

	for {
		select {
		case <-ctx.Done():
			logger.Info("Gracefully shutdown through context cancellation")
			return
		case <-ticker.C:
			maxRetries := 10
			now := time.Now().In(loc)

			startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 15, 0, 0, loc)
			endTime := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, loc)

			// 🔄 reset on new day
			currentDate := now.Format("02-Jan-2006")
			if lastTimeStampRecorded != currentDate {
				mu.Lock()
				*records = []models.ResponsePayload{}
				mu.Unlock()

				lastTimeStampRecorded = currentDate
				isWrittenToDB = false
				logger.Info("New trading day detected. Cleared previous records.")
			}

			if now.Before(startTime) {
				logger.Info("Market not started yet. Waiting for market to open.")
				continue
			} else if now.After(endTime) {
				logger.Info("Market closed. Stopping data fetch.")
				if !isWrittenToDB {
					err := db.WriteToDB(ctx, *records)
					if err != nil {
						logger.Error("Failed to write to database", slog.Any("error", err))
						continue
					}
					isWrittenToDB = true
				}
				continue
			}

			if isMarketHoliday(now, loc) {
				logger.Info("Market is closed today. Skipping data fetch.")
				continue
			}

			var newRecords models.Records
			for range maxRetries {
				var recordFetchError error
				newRecords, recordFetchError = api.FetchData(logger)
				if recordFetchError != nil {
					logger.Error("Failed to fetch data", slog.Any("error", recordFetchError))
				}
				datePart := strings.Split(newRecords.TimeStamp, " ")[0]
				if now.Format("02-Jan-2006") != datePart {
					newRecords = models.Records{}
				} else {
					break
				}
			}

			responsePayload := extractResponsePayload(newRecords)

			mu.Lock()
			*records = append(*records, responsePayload...)
			mu.Unlock()
		}
	}
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
