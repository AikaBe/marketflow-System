package postgres

import (
	"database/sql"
	"marketflow/internal/domain"
	"time"
)

func (a *Adapter) GetAggregatedPriceForSymbol(symbol string) (*domain.AggregatedResponse, error) {
	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`, symbol)

	var pair, exchange string
	var ts time.Time
	var avg, min, max float64

	err := row.Scan(&pair, &exchange, &ts, &avg, &min, &max)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &domain.AggregatedResponse{
		Pair:      pair,
		Exchange:  exchange,
		Timestamp: ts.Format(time.RFC3339),
		Avg:       avg,
		Min:       min,
		Max:       max,
	}, nil
}
