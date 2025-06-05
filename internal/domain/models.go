package domain

import "time"

type MarketData struct {
	Exchange string
	Symbol   string
	Price    float64
	Volume   float64
	Time     int64
}

type AggregatedMarketData struct {
	ID           int64     `db:"id"`
	PairName     string    `db:"pair_name"`
	Exchange     string    `db:"exchange"`
	Timestamp    time.Time `db:"timestamp"`
	AveragePrice float64   `db:"average_price"`
	MinPrice     float64   `db:"min_price"`
	MaxPrice     float64   `db:"max_price"`
}
