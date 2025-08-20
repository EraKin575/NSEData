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
)

func ProcessingOptionChain(ctx context.Context, records *[]models.ResponsePayload, db *db.DB, loc *time.Location, mu *sync.RWMutex) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	writtenToDB := false
	var lastTimeStampRecorded string

	for {
		now := time.Now().In(loc)

		startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 15, 0, 0, loc)
		endTime := time.Date(now.Year(), now.Month(), now.Day(), 15, 40, 0, 0, loc)
		resetTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 55, 0, 0, loc)

		switch {
		case now.Before(startTime):
			fmt.Println("Waiting for market to open:", now)
			time.Sleep(time.Until(startTime))
			continue

		case now.After(endTime):
			if !writtenToDB {
				err := db.WriteToDB(ctx, *records)
				if err != nil {
					log.Println("Error writing to DB:", err)
				}
				writtenToDB = true
				fmt.Println("Data written to DB at:", now)
			}

			if now.After(resetTime) {
				mu.Lock()
				*records = []models.ResponsePayload{}
				mu.Unlock()
				writtenToDB = false
				fmt.Println("Records reset at:", now)
			}

			continue
		}

		newRecords, err := api.FetchData()
		if err != nil {
			log.Println("records not fetched", err)
			continue
		}
		// Add safety checks and debugging
		if newRecords.TimeStamp == "" {
			fmt.Println("Empty timestamp received")
			continue
		}

		timestampParts := strings.Split(newRecords.TimeStamp, " ")
		if len(timestampParts) == 0 {
			fmt.Println("Invalid timestamp format:", newRecords.TimeStamp)
			continue
		}

		datePart := timestampParts[0]
		expectedDate := time.Now().In(loc).Format("02-Jan-2006")

		fmt.Printf("Timestamp: %s\n", newRecords.TimeStamp)
		fmt.Printf("Date part: %s\n", datePart)
		fmt.Printf("Expected: %s\n", expectedDate)
		fmt.Printf("Last recorded: %s\n", lastTimeStampRecorded)

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

		<-ticker.C
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

		timeStamp, err := time.Parse("2006-01-02 15:04:05", records.TimeStamp)
		if err != nil {
			timeStamp = time.Time{}
		}
		expiryDate, err := time.Parse("02-Jan-2006", record.ExpiryDate)
		if err != nil {
			expiryDate = time.Time{}
		}

		response = append(response, models.ResponsePayload{
			Timestamp:              timeStamp,
			ExpiryDate:             expiryDate,
			StrikePrice:            record.StrikePrice,
			UnderlyingValue:        records.UnderlyingValue,
			CEOpenInterest:         ceOI,
			CEChangeInOpenInterest: ceChOI,
			CETotalTradedVolume:    ceVol,
			CEImpliedVolatility:    ceIV,
			CELastPrice:            ceLTP,
			PEOpenInterest:         peOI,
			PEChangeInOpenInterest: peChOI,
			PETotalTradedVolume:    peVol,
			PEImpliedVolatility:    peIV,
			PELastPrice:            peLTP,
			PCR:                    pcr,
			IntraDayPCR:            intradayPCR,
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

func isFinite(value float64) bool {
	return !math.IsInf(value, 0) && !math.IsNaN(value)
}
