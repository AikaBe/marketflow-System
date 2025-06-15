package domain

type PriceUpdate struct {
	Symbol    string
	Price     float64
	Timestamp int64
	Exchange  string
}
