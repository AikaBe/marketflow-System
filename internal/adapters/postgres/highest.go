package postgres

import (
	"database/sql"
	"marketflow/internal/domain"
	"time"
)

func (a *ApiAdapter) GetHighestBySymbol(symbol string) (*domain.AggregatedResponse, error) {
	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1
		ORDER BY average_price DESC
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

func (a *ApiAdapter) GetHighestByExchange(exchange, symbol string) (*domain.AggregatedResponse, error) {
	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1 and exchange = $2
		ORDER BY average_price DESC
		LIMIT 1
	`, symbol, exchange)

	var pair string
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

func (a *ApiAdapter) QueryHighestPriceSince(symbol string, since time.Time) (*domain.AggregatedResponse, error) {
	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1 AND timestamp >= $2
	`, symbol, since)

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

func (a *ApiAdapter) QueryHighestSinceByExchange(exchange, symbol string, since time.Time) (*domain.AggregatedResponse, error) {
	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3
	`, symbol, exchange, since)

	var pair string
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
