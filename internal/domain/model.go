package domain

var TradingPairs = []string{
	"BTCUSDT",
	"ETHUSDT",
	"DOGEUSDT",
	"TONUSDT",
	"SOLUSDT",
}

var ExchangeNames = []string{
	"Exchange1",
	"Exchange2",
	"Exchange3",
}

type PriceUpdate struct {
	Symbol    string
	Price     float64
	Timestamp int64
	Exchange  string
}

type AggregatedResponse struct {
	Pair      string  `json:"pair"`
	Exchange  string  `json:"exchange"`
	Timestamp string  `json:"timestamp"`
	Avg       float64 `json:"avg"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
}
