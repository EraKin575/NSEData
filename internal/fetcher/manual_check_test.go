package fetcher

import (
	"context"
	"testing"
	"time"
)

func TestManualFetchOptionChain(t *testing.T) {
	b := NewBrowser()
	defer b.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chain, err := b.getOptionChain(ctx)
	if err != nil {
		t.Fatalf("getOptionChain error: %v", err)
	}

	t.Logf("timestamp=%q underlying=%v expiries=%v records=%d",
		chain.Records.TimeStamp, chain.Records.UnderlyingValue,
		chain.Records.ExpiryDates, len(chain.Records.Data))

	if len(chain.Records.Data) > 0 {
		d := chain.Records.Data[0]
		t.Logf("sample record: strike=%v expiry=%v CE=%+v PE=%+v", d.StrikePrice, d.ExpiryDate, d.CE, d.PE)
	}
}
