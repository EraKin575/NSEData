package processing

import (
	"context"
	"log/slog"
	"server/internal/db"
	"server/internal/models"
	"strings"
	"sync"
	"time"
)

type ProcessingService struct {
	Reader   Reader
	DBWriter DBWriter
}

func (r *ProcessingService) ProcessingOptionChain(ctx context.Context, db *db.DB, mu *sync.RWMutex, logger *slog.Logger, records *[]models.ResponsePayload) error {
	var lastTimeStampRecorded string
	isWrittenToDB := false

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		logger.Error("Failed to load location", slog.String("error", err.Error()))
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Gracefully shutdown through context cancellation")
			return nil
		default:
			maxRetries := 10
			now := time.Now().In(loc)

			startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 15, 0, 0, loc)
			endTime := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, loc)

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
					err := r.DBWriter.WriteToDB(ctx, records)
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
				if len(*records) == 0 {
					data, err := r.Reader.ReadStream(ctx)
					if err != nil {
						logger.Error("Failed to fetch data", slog.Any("error", err))
						continue
					}
					if len(data) == 0 {
						logger.Info("No data in stream. Retrying...")
						time.Sleep(10 * time.Second)
						continue
					}
				}
				newRecords, recordFetchError = r.Reader.ReadLatest(ctx)
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
