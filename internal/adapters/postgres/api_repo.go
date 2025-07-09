package postgres

import "database/sql"

type AggregatedAdapter struct {
	db *sql.DB
}

func NewAggregatedAdapter(connStr string) (*AggregatedAdapter, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &AggregatedAdapter{db: db}, nil
}

func (a *AggregatedAdapter) Close() error {
	return a.db.Close()
}
