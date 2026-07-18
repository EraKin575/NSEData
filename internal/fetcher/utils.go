package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"server/internal/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

const (
	userAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	optionChainPage   = "https://www.nseindia.com/option-chain"
	cookieRefreshTTL  = 10 * time.Minute
	navigationTimeout = 30 * time.Second
)

var (
	browserOnce   sync.Once
	browserCtx    context.Context
	browserCancel context.CancelFunc

	sessionMu sync.Mutex
	cookiesAt time.Time
)

// browser lazily starts a single long-lived headless Chrome instance and
// keeps it (and one tab) alive for the life of the process. Re-fetching
// NSE's session cookies requires a full page load, which is far too slow to
// redo on every poll tick (as often as once a second before market open), so
// the tab is reused and cookies are only refreshed periodically.
func browser() context.Context {
	browserOnce.Do(func() {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.UserAgent(userAgent),
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)
		if execPath := os.Getenv("CHROME_PATH"); execPath != "" {
			opts = append(opts, chromedp.ExecPath(execPath))
		}

		allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
		browserCtx, browserCancel = chromedp.NewContext(allocCtx)

		// The first Run on a fresh chromedp context allocates the browser and
		// binds its lifetime to whatever context that call used; cancelling a
		// context.WithTimeout wrapper afterwards would tear the whole browser
		// down (see the chromedp.Run docs). So bootstrap it here once, with
		// the bare uncancelled browserCtx, before any timeout-wrapped calls.
		_ = chromedp.Run(browserCtx)
	})
	return browserCtx
}

// CloseBrowser shuts down the shared headless Chrome instance. Should be
// called once on process shutdown.
func CloseBrowser() {
	if browserCancel != nil {
		browserCancel()
	}
}

// refreshSession loads the option-chain page in the shared tab so NSE issues
// its session cookies, staying on-origin so subsequent in-page fetch() calls
// carry them the same way the real site's own JS does. force bypasses the TTL,
// used when a fetch comes back looking like a bot-check page rather than JSON.
func refreshSession(force bool) error {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	if !force && time.Since(cookiesAt) < cookieRefreshTTL {
		return nil
	}

	ctx, cancel := context.WithTimeout(browser(), navigationTimeout)
	defer cancel()

	if err := chromedp.Run(ctx,
		chromedp.Navigate(optionChainPage),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	); err != nil {
		return fmt.Errorf("failed to set cookies: %w", err)
	}

	cookiesAt = time.Now()
	return nil
}

// fetchJSON runs a fetch() from inside the already-loaded nseindia.com page
// so the request carries the page's session cookies and headers exactly as
// the site's own frontend would send them, then unmarshals the JSON body.
func fetchJSON(url string, out any) error {
	ctx, cancel := context.WithTimeout(browser(), navigationTimeout)
	defer cancel()

	script := fmt.Sprintf(`fetch(%q, {headers: {"Accept": "application/json"}}).then(function(r) {
		if (!r.ok) { throw new Error("HTTP " + r.status); }
		return r.text();
	})`, url)

	var body string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &body, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	)
	if err != nil {
		return fmt.Errorf("in-page fetch of %s failed: %w", url, err)
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("empty response body from %s", url)
	}

	if err := json.Unmarshal([]byte(body), out); err != nil {
		// A bot-check/redirect page comes back as HTML, not JSON. Force a
		// cookie refresh on the next attempt rather than waiting out the TTL.
		sessionMu.Lock()
		cookiesAt = time.Time{}
		sessionMu.Unlock()

		preview := body
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return fmt.Errorf("json unmarshal failed for %s: %w | body preview: %s", url, err, preview)
	}

	return nil
}

func getOptionChain() (models.OptionChain, error) {
	if err := refreshSession(false); err != nil {
		return models.OptionChain{}, fmt.Errorf("cookie setup failed: %w", err)
	}

	var contractData struct {
		ExpiryDates []string `json:"expiryDates"`
	}
	contractInfoURL := "https://www.nseindia.com/api/option-chain-contract-info?symbol=NIFTY&type=Indices"
	if err := fetchJSON(contractInfoURL, &contractData); err != nil {
		return models.OptionChain{}, fmt.Errorf("failed to fetch contract info: %w", err)
	}

	if len(contractData.ExpiryDates) < 2 {
		return models.OptionChain{}, fmt.Errorf("not enough expiry dates available")
	}

	firstExpiry := contractData.ExpiryDates[0]
	secondExpiry := contractData.ExpiryDates[1]

	var allData []models.OptionData
	var expiryDates []string
	var timestamp string
	var underlyingValue float64
	seen := make(map[string]struct{})

	for _, expiry := range []string{firstExpiry, secondExpiry} {
		url := fmt.Sprintf("https://www.nseindia.com/api/option-chain-v3?type=Indices&symbol=NIFTY&expiry=%s", expiry)

		var optionData models.OptionChain
		if err := fetchJSON(url, &optionData); err != nil {
			return models.OptionChain{}, fmt.Errorf("failed to fetch data for expiry %s: %w", expiry, err)
		}

		// Each call is scoped to a single expiry, so its rows shouldn't
		// overlap with the other call's, but dedupe by contract identifier
		// (unique per symbol+expiry+type+strike) to be safe rather than
		// relying on that.
		for _, row := range optionData.Records.Data {
			key := row.ExpiryDate + "|" + strconv.FormatFloat(row.StrikePrice, 'f', -1, 64)
			if row.CE != nil {
				key = row.CE.Identifier
			} else if row.PE != nil {
				key = row.PE.Identifier
			}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			allData = append(allData, row)
		}

		// The contract-info/option-chain-v3 endpoints return the full expiry
		// catalog for the symbol regardless of which expiry was requested,
		// so it's identical across both calls here - only capture it once.
		if expiryDates == nil {
			expiryDates = optionData.Records.ExpiryDates
		}
		timestamp = optionData.Records.TimeStamp
		underlyingValue = optionData.Records.UnderlyingValue
	}

	return models.OptionChain{
		Records: models.Records{
			ExpiryDates:     expiryDates,
			Data:            allData,
			TimeStamp:       timestamp,
			UnderlyingValue: underlyingValue,
		},
	}, nil
}

func isMarketHoliday(date time.Time, loc *time.Location) bool {
	if date.In(loc).Weekday() == time.Saturday || date.In(loc).Weekday() == time.Sunday {
		return true
	}

	return false
}
