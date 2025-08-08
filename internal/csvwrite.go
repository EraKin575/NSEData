package csvwriter

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"server/models"
	"time"
)

func WriteRecordsToCSV(records []models.Records) {
	layout := "02-Jan-2006 15:04:05"
	type OptionWithMeta struct {
		Option     models.OptionData
		Spot       float64
		Timestamp  time.Time
		ExpiryDate string
	}

	expiryGroups := make(map[string][]OptionWithMeta)

	for _, snapshot := range records {
		for _, od := range snapshot.Data {
			expiry := od.ExpiryDate
			timeStamp, err := time.Parse(layout, snapshot.TimeStamp)
			if err != nil {
				log.Printf("❌ Failed to parse timestamp %s: %v", snapshot.TimeStamp, err)
				continue
			}
			expiryGroups[expiry] = append(expiryGroups[expiry], OptionWithMeta{
				Option:     od,
				Spot:       snapshot.UnderlyingValue,
				Timestamp:  timeStamp,
				ExpiryDate: expiry,
			})
		}
	}

	today := time.Now().Format("2006-01-02")

	for expiry, entries := range expiryGroups {
		filename := fmt.Sprintf("%s_%s.csv", today, expiry)
		file, err := os.Create(filename)
		if err != nil {
			log.Printf("❌ Failed to create file %s: %v", filename, err)
			continue
		}
		defer file.Close()

		writer := csv.NewWriter(file)

		// CSV Header as per your required order
		writer.Write([]string{
			"Timestamp",
			"Expiry Date",
			"COI",
			"CCOI",
			"CCOI%",
			"CVol",
			"CIV",
			"CE LTP",
			"Spot",
			"Strike Price",
			"PE LTP",
			"PE IV",
			"PE Vol",
			"PCOI%",
			"PCOI",
			"POI",
			"IntraDay PCR",
			"PCR",
		})

		for _, entry := range entries {
			od := entry.Option
			spot := entry.Spot
			timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
			expiry := entry.ExpiryDate

			ce := od.CE
			pe := od.PE

			// CE values
			ceOI, ceChOI, ceVol, ceIV, ceLTP := 0.0, 0.0, 0, 0.0, 0.0
			if ce != nil {
				ceOI = ce.OpenInterest
				ceChOI = ce.ChangeInOpenInterest
				ceVol = ce.TotalTradedVolume
				ceIV = ce.ImpliedVolatility
				ceLTP = ce.LastPrice
			}

			// PE values
			peOI, peChOI, peVol, peIV, peLTP := 0.0, 0.0, 0, 0.0, 0.0
			if pe != nil {
				peOI = pe.OpenInterest
				peChOI = pe.ChangeInOpenInterest
				peVol = pe.TotalTradedVolume
				peIV = pe.ImpliedVolatility
				peLTP = pe.LastPrice
			}

			// Calculated values
			ccoiPct := calculatePercentage(ceChOI, ceOI)
			pcoiPct := calculatePercentage(peChOI, peOI)
			pcr := calculateRatio(peOI, ceOI)
			intradayPCR := calculateRatio(peChOI, ceChOI)

			row := []string{
				timestamp,
				expiry,
				fmt.Sprintf("%.0f", ceOI),
				fmt.Sprintf("%.0f", ceChOI),
				ccoiPct,
				fmt.Sprintf("%d", ceVol),
				fmt.Sprintf("%.2f", ceIV),
				fmt.Sprintf("%.2f", ceLTP),
				fmt.Sprintf("%.2f", spot),
				fmt.Sprintf("%.2f", od.StrikePrice),
				fmt.Sprintf("%.2f", peLTP),
				fmt.Sprintf("%.2f", peIV),
				fmt.Sprintf("%d", peVol),
				pcoiPct,
				fmt.Sprintf("%.0f", peChOI),
				fmt.Sprintf("%.0f", peOI),
				intradayPCR,
				pcr,
			}

			writer.Write(row)
		}

		writer.Flush()

		if err := writer.Error(); err != nil {
			log.Printf("❌ Error writing to CSV %s: %v", filename, err)
		} else {
			log.Printf("✅ CSV written: %s (%d rows)", filename, len(entries))
		}
	}
}

func calculatePercentage(change, total float64) string {
	if total == 0 || !isFinite(change/total) {
		return "-"
	}
	return fmt.Sprintf("%.2f", (change/total)*100)
}

func calculateRatio(num, denom float64) string {
	if denom == 0 || !isFinite(num/denom) {
		return "-"
	}
	return fmt.Sprintf("%.2f", num/denom)
}

func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}
