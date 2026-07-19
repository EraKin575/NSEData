package csvexport

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"server/internal/models"
	"strconv"
	"time"
)

var header = []string{
	"timestamp", "expiry_date", "strike_price", "underlying_value",
	"ce_oi", "ce_ch_oi", "ce_ch_oi_pct", "ce_vol", "ce_iv", "ce_ltp",
	"pe_oi", "pe_ch_oi", "pe_ch_oi_pct", "pe_vol", "pe_iv", "pe_ltp",
	"intraday_pcr", "pcr",
}

// ToCSV renders option chain records as CSV bytes.
func ToCSV(records []models.ResponsePayload) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write csv header: %w", err)
	}

	for _, p := range records {
		row := []string{
			p.Timestamp.Format(time.RFC3339),
			p.ExpiryDate.Format("2006-01-02"),
			formatFloat(p.StrikePrice),
			formatFloat(p.UnderlyingValue),
			formatFloat(p.CEOpenInterest),
			formatFloat(p.CEChangeInOpenInterest),
			formatFloat(p.CEChangeInOpenInterestPercentage),
			strconv.Itoa(p.CETotalTradedVolume),
			formatFloat(p.CEImpliedVolatility),
			formatFloat(p.CELastPrice),
			formatFloat(p.PEOpenInterest),
			formatFloat(p.PEChangeInOpenInterest),
			formatFloat(p.PEChangeInOpenInterestPercentage),
			strconv.Itoa(p.PETotalTradedVolume),
			formatFloat(p.PEImpliedVolatility),
			formatFloat(p.PELastPrice),
			formatFloat(p.IntraDayPCR),
			formatFloat(p.PCR),
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv writer error: %w", err)
	}

	return buf.Bytes(), nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
