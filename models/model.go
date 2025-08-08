package models

type Option struct {
	StrikePrice           float64 `json:"strikePrice"`
	ExpiryDate            string  `json:"expiryDate"`
	Underlying            string  `json:"underlying"`
	Identifier            string  `json:"identifier"`
	OpenInterest          float64 `json:"openInterest"`
	ChangeInOpenInterest  float64 `json:"changeinOpenInterest"`
	PChangeInOpenInterest float64 `json:"pchangeinOpenInterest"`
	TotalTradedVolume     int     `json:"totalTradedVolume"`
	ImpliedVolatility     float64 `json:"impliedVolatility"`
	LastPrice             float64 `json:"lastPrice"`
	Change                float64 `json:"change"`
	PChange               float64 `json:"pChange"`
	TotalBuyQuantity      int     `json:"totalBuyQuantity"`
	TotalSellQuantity     int     `json:"totalSellQuantity"`
	BidQty                int     `json:"bidQty"`
	BidPrice              float64 `json:"bidprice"`
	AskQty                int     `json:"askQty"`
	AskPrice              float64 `json:"askPrice"`
	UnderlyingValue       float64 `json:"underlyingValue"`
}

type OptionData struct {
	StrikePrice float64 `json:"strikePrice"`
	ExpiryDate  string  `json:"expiryDate"`
	CE          *Option `json:"CE,omitempty"`
	PE          *Option `json:"PE,omitempty"`
}

type Records struct {
	ExpiryDates     []string     `json:"expiryDates"`
	Data            []OptionData `json:"data"`
	TimeStamp       string       `json:"timestamp"`
	UnderlyingValue float64      `json:"underlyingValue"`
}

type OptionChain struct {
	Records Records `json:"records"`
}
