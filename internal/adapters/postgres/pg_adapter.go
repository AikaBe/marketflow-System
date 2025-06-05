package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"marketflow/internal/domain"
)

type PostgresAdapter struct {
	db *sql.DB
}

func NewPostgresAdapter(connStr string) (*PostgresAdapter, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &PostgresAdapter{db: db}, nil
}

func (p *PostgresAdapter) Close() error {
	return p.db.Close()
}

// BatchInsertAggregatedData вставляет несколько записей в транзакции
func (p *PostgresAdapter) BatchInsertAggregatedData(ctx context.Context, data []domain.AggregatedMarketData) error {
	if len(data) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO aggregated_prices (pair_name, exchange, timestamp, average_price, min_price, max_price)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range data {
		_, err := stmt.ExecContext(ctx,
			d.PairName, d.Exchange, d.Timestamp, d.AveragePrice, d.MinPrice, d.MaxPrice,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to exec statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

// GetLatestPrice возвращает последний агрегированный прайс для пары и биржи
func (p *PostgresAdapter) GetLatestPrice(ctx context.Context, pairName, exchange string) (*domain.AggregatedMarketData, error) {
	sqlQuery := `
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM aggregated_prices
		WHERE pair_name = $1 AND exchange = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`

	row := p.db.QueryRowContext(ctx, sqlQuery, pairName, exchange)

	var res domain.AggregatedMarketData
	err := row.Scan(&res.PairName, &res.Exchange, &res.Timestamp, &res.AveragePrice, &res.MinPrice, &res.MaxPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &res, nil
}
