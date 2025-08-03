package utils

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
	expiryGroups := make(map[string][]models.OptionData)

	for _, snapshot := range records {
		for _, od := range snapshot.Data {
			expiry := od.ExpiryDate
			expiryGroups[expiry] = append(expiryGroups[expiry], od)
		}
	}

	today := time.Now().Format("2006-01-02")

	for expiry, data := range expiryGroups {
		filename := fmt.Sprintf("%s_%s.csv", today, expiry)
		file, err := os.Create(filename)
		if err != nil {
			log.Printf("❌ Failed to create file %s: %v", filename, err)
			continue
		}
		defer file.Close()

		writer := csv.NewWriter(file)

		// Header
		writer.Write([]string{
			"StrikePrice",
			"CE_OI", "CE_ChangeInOI", "CE_Volume", "CE_IV", "CE_LTP",
			"PE_OI", "PE_ChangeInOI", "PE_Volume", "PE_IV", "PE_LTP",
			"PCR", "IntradayPCR",
		})

		for _, od := range data {
			ce := od.CE
			pe := od.PE

			// Safe dereference
			ceOI, ceChOI, ceVol, ceIV, ceLTP := 0.0, 0.0, 0, 0.0, 0.0
			if ce != nil {
				ceOI = ce.OpenInterest
				ceChOI = ce.ChangeInOpenInterest
				ceVol = ce.TotalTradedVolume
				ceIV = ce.ImpliedVolatility
				ceLTP = ce.LastPrice
			}

			peOI, peChOI, peVol, peIV, peLTP := 0.0, 0.0, 0, 0.0, 0.0
			if pe != nil {
				peOI = pe.OpenInterest
				peChOI = pe.ChangeInOpenInterest
				peVol = pe.TotalTradedVolume
				peIV = pe.ImpliedVolatility
				peLTP = pe.LastPrice
			}

			pcr := calculateRatio(peOI, ceOI)
			intradayPCR := calculateRatio(peChOI, ceChOI)

			row := []string{
				fmt.Sprintf("%.2f", od.StrikePrice),

				fmt.Sprintf("%.0f", ceOI),
				fmt.Sprintf("%.0f", ceChOI),
				fmt.Sprintf("%d", ceVol),
				fmt.Sprintf("%.2f", ceIV),
				fmt.Sprintf("%.2f", ceLTP),

				fmt.Sprintf("%.0f", peOI),
				fmt.Sprintf("%.0f", peChOI),
				fmt.Sprintf("%d", peVol),
				fmt.Sprintf("%.2f", peIV),
				fmt.Sprintf("%.2f", peLTP),

				pcr,
				intradayPCR,
			}

			writer.Write(row)
		}

		writer.Flush()

		if err := writer.Error(); err != nil {
			log.Printf("❌ Error writing to CSV %s: %v", filename, err)
		} else {
			log.Printf("✅ CSV written: %s (%d rows)", filename, len(data))
		}
	}
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
