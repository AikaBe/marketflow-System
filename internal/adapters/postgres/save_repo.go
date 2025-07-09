package postgres

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

type Adapter struct {
	db *sql.DB
}

func NewPostgresAdapter(connStr string) (*Adapter, error) {
	slog.Info("Connecting to PostgreSQL", "dsn", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("Failed to open PostgreSQL connection", "err", err)
		return nil, err
	}
	slog.Info("Successfully connected to PostgreSQL")
	return &Adapter{db: db}, nil
}

func (a *Adapter) Close() error {
	slog.Info("Closing PostgreSQL connection")
	return a.db.Close()
}

func (a *Adapter) SaveAggregatedPrice(ctx context.Context, pair, exchange string, ts time.Time, avg, min, max float64) error {
	slog.Info("Saving aggregated price",
		"pair", pair,
		"exchange", exchange,
		"timestamp", ts,
		"avg", avg,
		"min", min,
		"max", max,
	)

	_, err := a.db.ExecContext(ctx, `
		INSERT INTO aggregated_prices (pair_name, exchange, timestamp, average_price, min_price, max_price)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, pair, exchange, ts, avg, min, max)
	if err != nil {
		slog.Error("Failed to save aggregated price",
			"pair", pair,
			"exchange", exchange,
			"timestamp", ts,
			"err", err,
		)
		return err
	}

	slog.Info("Successfully saved aggregated price", "pair", pair, "exchange", exchange)
	return nil
}
