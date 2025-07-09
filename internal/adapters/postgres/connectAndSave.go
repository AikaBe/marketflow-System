package postgres

import (
	"context"
	"database/sql"
	"time"
)

type Adapter struct {
	db *sql.DB
}

func NewPostgresAdapter(connStr string) (*Adapter, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &Adapter{db: db}, nil
}

func (a *Adapter) Close() error {
	return a.db.Close()
}

func (a *Adapter) SaveAggregatedPrice(ctx context.Context, pair, exchange string, ts time.Time, avg, min, max float64) error {
	_, err := a.db.ExecContext(ctx, `
		INSERT INTO aggregated_prices (pair_name, exchange, timestamp, average_price, min_price, max_price)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, pair, exchange, ts, avg, min, max)
	return err
}
