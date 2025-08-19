package processing

import (
	"fmt"
	"log"
	csvwriter "server/internal"
	"server/internal/api"
	"server/models"
	"sync"
	"time"
)

func ProcessingOptionChain(records *[]models.Records, loc *time.Location, mu *sync.RWMutex) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	csvWritten := false

	for {
		now := time.Now().In(loc)

		startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 16, 0, 0, loc)
		endTime := time.Date(now.Year(), now.Month(), now.Day(), 15, 31, 0, 0, loc)
		resetTime := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc)

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

		mu.Lock()
		*records = append(*records, newRecords)
		mu.Unlock()

		<-ticker.C
	}
}
