package models

import "time"

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

type ResponsePayload struct {
	Timestamp                        time.Time `json:"timestamp"`  // Snapshot timestamp
	ExpiryDate                       time.Time `json:"expiryDate"` // Expiry date of the option
	StrikePrice                      float64   `json:"strikePrice"`
	UnderlyingValue                  float64   `json:"underlyingValue"` // Current underlying (NIFTY) value
	CEOpenInterest                   float64   `json:"ceOpenInterest"`
	CEChangeInOpenInterestPercentage float64   `json:"ceOpenInterestPercentage"`
	CEChangeInOpenInterest           float64   `json:"ceChangeInOpenInterest"`
	CETotalTradedVolume              int       `json:"ceTotalTradedVolume"`
	CEImpliedVolatility              float64   `json:"ceImpliedVolatility"`
	CELastPrice                      float64   `json:"ceLastPrice"`
	PEOpenInterest                   float64   `json:"peOpenInterest"`
	PEChangeInOpenInterestPercentage float64   `json:"peOpenInterestPercentage"`
	PEChangeInOpenInterest           float64   `json:"peChangeInOpenInterest"`
	PETotalTradedVolume              int       `json:"peTotalTradedVolume"`
	PEImpliedVolatility              float64   `json:"peImpliedVolatility"`
	PELastPrice                      float64   `json:"peLastPrice"`
	IntraDayPCR                      float64   `json:"intraDayPCR"` // Change in PE OI / Change in CE OI
	PCR                              float64   `json:"pcr"`         // Total PE OI / Total CE OI

}
