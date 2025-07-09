package postgres

import (
	"database/sql"
	"log/slog"
	"marketflow/internal/domain"
	"time"
)

func (a *ApiAdapter) GetAvgBySymbol(symbol string) (*domain.AggregatedResponse, error) {
	slog.Info("Querying average price by symbol", "symbol", symbol)

	row := a.db.QueryRow(`
		SELECT AVG(average_price)
		FROM aggregated_prices
		WHERE pair_name = $1
	`, symbol)

	var avg float64
	err := row.Scan(&avg)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("No average price found for symbol", "symbol", symbol)
			return nil, nil
		}
		slog.Error("Failed to scan average price", "symbol", symbol, "err", err)
		return nil, err
	}

	slog.Info("Average price retrieved", "symbol", symbol, "avg", avg)

	return &domain.AggregatedResponse{
		Pair:      symbol,
		Timestamp: time.Now().Format(time.RFC3339),
		Avg:       avg,
	}, nil
}

func (a *ApiAdapter) GetAvgByExchange(exchange, symbol string) (*domain.AggregatedResponse, error) {
	slog.Info("Querying average price by exchange and symbol", "exchange", exchange, "symbol", symbol)

	row := a.db.QueryRow(`
		SELECT AVG(average_price)
		FROM aggregated_prices
		WHERE pair_name = $1 AND exchange = $2
	`, symbol, exchange)

	var avg float64
	err := row.Scan(&avg)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("No average price found for exchange and symbol", "exchange", exchange, "symbol", symbol)
			return nil, nil
		}
		slog.Error("Failed to scan average price", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}

	slog.Info("Average price retrieved", "exchange", exchange, "symbol", symbol, "avg", avg)

	return &domain.AggregatedResponse{
		Pair:      symbol,
		Exchange:  exchange,
		Timestamp: time.Now().Format(time.RFC3339),
		Avg:       avg,
	}, nil
}

func (a *ApiAdapter) QueryAvgSinceByExchange(exchange, symbol string, since time.Time) (*domain.AggregatedResponse, error) {
	slog.Info("Querying average price by exchange, symbol, and period", "exchange", exchange, "symbol", symbol, "since", since.Format(time.RFC3339))

	row := a.db.QueryRow(`
		SELECT AVG(average_price)
		FROM aggregated_prices
		WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3
	`, symbol, exchange, since)

	var avg float64
	err := row.Scan(&avg)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("No average price found in time range", "exchange", exchange, "symbol", symbol, "since", since)
			return nil, nil
		}
		slog.Error("Failed to scan average price", "exchange", exchange, "symbol", symbol, "since", since, "err", err)
		return nil, err
	}

	slog.Info("Average price retrieved for period", "exchange", exchange, "symbol", symbol, "avg", avg)

	return &domain.AggregatedResponse{
		Pair:      symbol,
		Exchange:  exchange,
		Timestamp: time.Now().Format(time.RFC3339),
		Avg:       avg,
	}, nil
}
