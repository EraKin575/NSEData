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

/*
ProcessingOptionChain fetches option chain data at regular intervals,
processes it, and writes it to CSV files.
It runs in a loop, checking the current time against market open and close times.
It uses a ticker to fetch data every 3 minutes.

*/

func ProcessingOptionChain(records *[]models.Records, startTime, endTime time.Time, loc *time.Location, mu *sync.RWMutex) {

	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		now := time.Now().In(loc)

		switch {
		case now.Before(startTime):
			fmt.Print("waiting for market to open", time.Now())
			wait := time.Until(startTime)
			time.Sleep(wait)
			continue

		case now.After(endTime):
			csvwriter.WriteRecordsToCSV(*records)
			*records = []models.Records{}

			wait := time.Until(startTime)
			time.Sleep(wait)
			continue
		}

		newRecords, err := api.FetchData()
		if err != nil {
			log.Fatal("records not fetched", err)
		}

		mu.Lock()

		*records = append(*records, newRecords)

		mu.Unlock()

		<-ticker.C

	}
}
