package processing

import (
	"math"
	"server/internal/models"
	"time"
)

func isMarketHoliday(date time.Time, loc *time.Location) bool {
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return true
	}

	return false
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
