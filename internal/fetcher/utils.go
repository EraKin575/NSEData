package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"server/internal/models"
	"strings"
	"time"
)

var (
	headers = map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Referer":    "https://www.nseindia.com/option-chain",
		"Accept":     "application/json, text/plain, */*",
	}
	client *http.Client
)

func init() {
	jar, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
	}
}

func setCookies() error {
	req, err := http.NewRequest("GET", "https://www.nseindia.com/option-chain", nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set cookies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cookie setup: NSE returned HTTP %d %s", resp.StatusCode, resp.Status)
	}

	_, _ = io.ReadAll(resp.Body) // needed to populate cookies
	return nil
}

func getOptionChain() (models.OptionChain, error) {
	if err := setCookies(); err != nil {
		return models.OptionChain{}, fmt.Errorf("cookie setup failed: %w", err)
	}

	contractInfoURL := "https://www.nseindia.com/api/option-chain-contract-info?symbol=NIFTY&type=Indices"
	contractReq, err := http.NewRequest("GET", contractInfoURL, nil)
	if err != nil {
		return models.OptionChain{}, err
	}
	for k, v := range headers {
		contractReq.Header.Set(k, v)
	}

	contractResp, err := client.Do(contractReq)
	if err != nil {
		return models.OptionChain{}, fmt.Errorf("failed to fetch contract info: %w", err)
	}
	defer contractResp.Body.Close()

	if contractResp.StatusCode != http.StatusOK {
		return models.OptionChain{}, fmt.Errorf("contract info: NSE returned HTTP %d %s", contractResp.StatusCode, contractResp.Status)
	}

	contractBody, err := io.ReadAll(contractResp.Body)
	if err != nil {
		return models.OptionChain{}, err
	}

	var contractData struct {
		ExpiryDates []string `json:"expiryDates"`
	}
	if err := json.Unmarshal(contractBody, &contractData); err != nil {
		return models.OptionChain{}, fmt.Errorf("failed to parse contract info: %w", err)
	}

	if len(contractData.ExpiryDates) == 0 {
		return models.OptionChain{}, fmt.Errorf("no expiry dates available")
	}

	firstExpiry := contractData.ExpiryDates[0]
	url := fmt.Sprintf("https://www.nseindia.com/api/option-chain-v3?type=Indices&symbol=NIFTY&expiry=%s", firstExpiry)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.OptionChain{}, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return models.OptionChain{}, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.OptionChain{}, fmt.Errorf("NSE returned HTTP %d %s", resp.StatusCode, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return models.OptionChain{}, fmt.Errorf("unexpected content-type: %s", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.OptionChain{}, err
	}

	var optionData models.OptionChain
	if err := json.Unmarshal(body, &optionData); err != nil {
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return models.OptionChain{}, fmt.Errorf("json unmarshal failed: %w | body preview: %s", err, preview)
	}
	return optionData, nil
}

func isMarketHoliday(date time.Time, loc *time.Location) bool {
	if date.In(loc).Weekday() == time.Saturday || date.In(loc).Weekday() == time.Sunday {
		return true
	}

	return false
}
