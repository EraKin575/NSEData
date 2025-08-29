package processing

import (
	"context"
	"fmt"
	"log"
	"math"
	"server/internal/api"
	"server/internal/db"
	"server/models"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

func ProcessingOptionChain(ctx context.Context, records *[]models.ResponsePayload, db *db.DB, loc *time.Location, mu *sync.RWMutex) {
	c := cron.New(
		cron.WithLocation(loc),
		cron.WithSeconds(), // allow seconds field in cron expression
	)

	writtenToDB := false
	var lastTimeStampRecorded string

	// Run every 3 minutes starting at 09:15 up to 15:39, Mon–Fri
	_, err := c.AddFunc("0 15-39/3 9-15 * * MON-FRI", func() {
		now := time.Now().In(loc)
		resetTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 55, 0, 0, loc)

		// Fetch new records
		newRecords, err := api.FetchData()
		if err != nil {
			log.Println("records not fetched:", err)
			return
		}
		if newRecords.TimeStamp == "" {
			fmt.Println("Empty timestamp received")
			return
		}

		timestampParts := strings.Split(newRecords.TimeStamp, " ")
		if len(timestampParts) == 0 {
			fmt.Println("Invalid timestamp format:", newRecords.TimeStamp)
			return
		}

		datePart := timestampParts[0]
		expectedDate := now.Format("02-Jan-2006")

		// Only append new records if timestamp is unique for the same day
		if datePart == expectedDate && lastTimeStampRecorded != newRecords.TimeStamp {
			mu.Lock()
			*records = append(*records, extractResponsePayload(newRecords)...)
			mu.Unlock()
			lastTimeStampRecorded = newRecords.TimeStamp
			fmt.Printf("Record added! Total: %d\n", len(*records))
		} else {
			fmt.Printf("Record rejected - Date match: %t, Timestamp different: %t\n",
				datePart == expectedDate,
				lastTimeStampRecorded != newRecords.TimeStamp)
		}

		// Write once after market close
		if !writtenToDB && now.Hour() == 15 && now.Minute() >= 40 {
			if err := db.WriteToDB(ctx, *records); err != nil {
				log.Println("Error writing to DB:", err)
			}
			writtenToDB = true
			fmt.Println("Data written to DB at:", now)
		}

		// Reset records late at night
		if now.After(resetTime) {
			mu.Lock()
			*records = []models.ResponsePayload{}
			mu.Unlock()
			writtenToDB = false
			lastTimeStampRecorded = "" // also reset last timestamp
			fmt.Println("Records reset at:", now)
		}
	})

	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}

	c.Start()
	<-ctx.Done()
	fmt.Println("Processing stopped")
	c.Stop()
}

// --- helpers ---

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

		// Calculate metrics
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
