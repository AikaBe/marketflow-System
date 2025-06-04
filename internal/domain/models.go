package domain

type MarketData struct {
	Exchange string
	Symbol   string
	Price    float64
	Volume   float64
	Time     int64
}
