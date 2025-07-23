package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"server/models"
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

	_, _ = io.ReadAll(resp.Body) // Needed to populate cookies
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

func FetchData() (models.Records, error) {
	chain, err := getOptionChain()
	if err != nil {
		return models.Records{}, err
	}

	if len(chain.Records.ExpiryDates) < 2 {
		return models.Records{}, fmt.Errorf("not enough expiry dates")
	}

	firstExpiry := chain.Records.ExpiryDates[0]
	secondExpiry := chain.Records.ExpiryDates[1]

	var expiryOneResult, expiryTwoResult []models.OptionData
	for _, entry := range chain.Records.Data {
		data := models.OptionData{
			StrikePrice: entry.StrikePrice,
			CE:          entry.CE,
			PE:          entry.PE,
			ExpiryDate:  entry.ExpiryDate,
		}
		switch entry.ExpiryDate {
		case firstExpiry:
			expiryOneResult = append(expiryOneResult, data)
		case secondExpiry:
			expiryTwoResult = append(expiryTwoResult, data)
		}
	}

	result := append(expiryOneResult, expiryTwoResult...)
	return models.Records{
		ExpiryDates:     []string{firstExpiry, secondExpiry},
		Data:            result,
		TimeStamp:       time.Now().Format(time.RFC3339),
		UnderlyingValue: chain.Records.UnderlyingValue,
	}, nil
}
