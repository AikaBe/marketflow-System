package postgres

import (
	"database/sql"
	"log/slog"
	"time"

	"marketflow/internal/domain"
)

func (a *ApiAdapter) GetLowestBySymbol(symbol string) (*domain.AggregatedResponse, error) {
	slog.Info("Querying lowest price by symbol", "symbol", symbol)

	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1
		ORDER BY average_price ASC
		LIMIT 1
	`, symbol)

	var pair, exchange string
	var ts time.Time
	var avg, min, max float64

	err := row.Scan(&pair, &exchange, &ts, &avg, &min, &max)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("No lowest price found for symbol", "symbol", symbol)
			return nil, nil
		}
		slog.Error("Failed to scan lowest price by symbol", "symbol", symbol, "err", err)
		return nil, err
	}

	slog.Info("Lowest price retrieved", "symbol", symbol, "avg", avg)
	return &domain.AggregatedResponse{
		Pair:      pair,
		Exchange:  exchange,
		Timestamp: ts.Format(time.RFC3339),
		Avg:       avg,
		Min:       min,
		Max:       max,
	}, nil
}

func (a *ApiAdapter) GetLowestByExchange(exchange, symbol string) (*domain.AggregatedResponse, error) {
	slog.Info("Querying lowest price by exchange and symbol", "exchange", exchange, "symbol", symbol)

	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1 AND exchange = $2
		ORDER BY average_price ASC
		LIMIT 1
	`, symbol, exchange)

	var pair string
	var ts time.Time
	var avg, min, max float64

	err := row.Scan(&pair, &exchange, &ts, &avg, &min, &max)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("No lowest price found for exchange and symbol", "exchange", exchange, "symbol", symbol)
			return nil, nil
		}
		slog.Error("Failed to scan lowest price by exchange", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}

	slog.Info("Lowest price retrieved", "exchange", exchange, "symbol", symbol, "avg", avg)
	return &domain.AggregatedResponse{
		Pair:      pair,
		Exchange:  exchange,
		Timestamp: ts.Format(time.RFC3339),
		Avg:       avg,
		Min:       min,
		Max:       max,
	}, nil
}

func (a *ApiAdapter) QueryLowestPriceSince(symbol string, since time.Time) (*domain.AggregatedResponse, error) {
	slog.Info("Querying lowest price by symbol since time", "symbol", symbol, "since", since)

	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1 AND timestamp >= $2
		ORDER BY average_price ASC
		LIMIT 1
	`, symbol, since)

	var pair, exchange string
	var ts time.Time
	var avg, min, max float64

	err := row.Scan(&pair, &exchange, &ts, &avg, &min, &max)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("No lowest price found for symbol in period", "symbol", symbol, "since", since)
			return nil, nil
		}
		slog.Error("Failed to scan lowest price by symbol since", "symbol", symbol, "since", since, "err", err)
		return nil, err
	}

	slog.Info("Lowest price retrieved", "symbol", symbol, "since", since, "avg", avg)
	return &domain.AggregatedResponse{
		Pair:      pair,
		Exchange:  exchange,
		Timestamp: ts.Format(time.RFC3339),
		Avg:       avg,
		Min:       min,
		Max:       max,
	}, nil
}

func (a *ApiAdapter) QueryLowestSinceByExchange(exchange, symbol string, since time.Time) (*domain.AggregatedResponse, error) {
	slog.Info("Querying lowest price by exchange and symbol since", "exchange", exchange, "symbol", symbol, "since", since)

	row := a.db.QueryRow(`
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3
		ORDER BY average_price ASC
		LIMIT 1
	`, symbol, exchange, since)

	var pair string
	var ts time.Time
	var avg, min, max float64

	err := row.Scan(&pair, &exchange, &ts, &avg, &min, &max)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("No lowest price found for exchange, symbol, and period", "exchange", exchange, "symbol", symbol, "since", since)
			return nil, nil
		}
		slog.Error("Failed to scan lowest price by exchange and symbol since", "exchange", exchange, "symbol", symbol, "since", since, "err", err)
		return nil, err
	}

	slog.Info("Lowest price retrieved", "exchange", exchange, "symbol", symbol, "since", since, "avg", avg)
	return &domain.AggregatedResponse{
		Pair:      pair,
		Exchange:  exchange,
		Timestamp: ts.Format(time.RFC3339),
		Avg:       avg,
		Min:       min,
		Max:       max,
	}, nil
}
