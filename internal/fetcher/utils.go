package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"server/internal/models"
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

	_, _ = io.ReadAll(resp.Body) // needed to populate cookies
	return nil
}

func getOptionChain() (models.OptionChain, error) {
	if err := setCookies(); err != nil {
		return models.OptionChain{}, fmt.Errorf("cookie setup failed: %w", err)
	}

	url := "https://www.nseindia.com/api/option-chain-indices?symbol=NIFTY"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.OptionChain{}, err
	}

	var optionData models.OptionChain
	if err := json.Unmarshal(body, &optionData); err != nil {
		return models.OptionChain{}, err
	}
	return optionData, nil
}

func isMarketHoliday(date time.Time, loc *time.Location) bool {
	if date.In(loc).Weekday() == time.Saturday || date.In(loc).Weekday() == time.Sunday {
		return true
	}

	return false
}
