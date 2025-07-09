package postgres

import (
	"database/sql"
	"log/slog"
	"marketflow/internal/domain"
	"time"
)

func (a *ApiAdapter) GetPriceForSymbol(symbol string) (*domain.AggregatedResponse, error) {
	slog.Info("Querying latest price for symbol", "symbol", symbol)

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
			slog.Warn("No latest price found for symbol", "symbol", symbol)
			return nil, nil
		}
		slog.Error("Failed to scan latest price for symbol", "symbol", symbol, "err", err)
		return nil, err
	}

	slog.Info("Latest price retrieved", "symbol", symbol, "avg", avg, "timestamp", ts)

	return &domain.AggregatedResponse{
		Pair:      pair,
		Exchange:  exchange,
		Timestamp: ts.Format(time.RFC3339),
		Avg:       avg,
		Min:       min,
		Max:       max,
	}, nil
}

func (a *ApiAdapter) GetPriceForExchange(exchange, symbol string) (*domain.AggregatedResponse, error) {
	slog.Info("Querying latest price for symbol by exchange", "exchange", exchange, "symbol", symbol)

	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1 AND exchange = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`, symbol, exchange)

	var pair string
	var ts time.Time
	var avg, min, max float64

	err := row.Scan(&pair, &exchange, &ts, &avg, &min, &max)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("No latest price found for symbol and exchange", "symbol", symbol, "exchange", exchange)
			return nil, nil
		}
		slog.Error("Failed to scan latest price", "symbol", symbol, "exchange", exchange, "err", err)
		return nil, err
	}

	slog.Info("Latest price retrieved", "symbol", symbol, "exchange", exchange, "avg", avg, "timestamp", ts)

	return &domain.AggregatedResponse{
		Pair:      pair,
		Exchange:  exchange,
		Timestamp: ts.Format(time.RFC3339),
		Avg:       avg,
		Min:       min,
		Max:       max,
	}, nil
}
