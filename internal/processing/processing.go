package processing

import (
	"fmt"
	"log"
	csvwriter "server/internal"
	"server/internal/api"
	"server/models"
	"strings"
	"sync"
	"time"
)

func ProcessingOptionChain(records *[]models.Records, loc *time.Location, mu *sync.RWMutex) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	csvWritten := false
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
			if !csvWritten {
				csvwriter.WriteRecordsToCSV(*records)
				csvWritten = true
				fmt.Println("CSV written at:", now)
			}

			if now.After(resetTime) {
				mu.Lock()
				*records = []models.Records{}
				mu.Unlock()
				csvWritten = false
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
			*records = append(*records, newRecords)
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
