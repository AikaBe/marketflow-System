package postgres

import "database/sql"

type ApiAdapter struct {
	db *sql.DB
}

func NewApiAdapter(connStr string) (*ApiAdapter, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &ApiAdapter{db: db}, nil
}

func (a *ApiAdapter) Close() error {
	return a.db.Close()
}

func (a *ApiAdapter) Ping() error {
	return a.db.Ping()
}
