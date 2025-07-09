package api

import (
	"errors"
	"marketflow/internal/domain"
	"strings"
	"time"
)

func (s *APIService) GetHighestBySymbol(symbol string) (*domain.AggregatedResponse, error) {
	symbol = strings.ToUpper(symbol)

	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}

	data, err := s.repo.GetHighestBySymbol(symbol)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *APIService) GetHighestByExchange(path string) (*domain.AggregatedResponse, error) {
	parts := strings.Split(path, "/")
	exchange := parts[0]
	symbol := parts[1]
	symbol = strings.ToUpper(symbol)
	data, err := s.repo.GetHighestByExchange(exchange, symbol)
	if err != nil {
		return nil, err
	}

	return data, err
}

func (s *APIService) GetHighestByPeriod(symbol string, period time.Duration) (*domain.AggregatedResponse, error) {
	since := time.Now().Add(-period)

	return s.repo.QueryHighestPriceSince(symbol, since)
}

func (s *APIService) QueryHighestSinceByExchange(exchange, symbol string, period time.Duration) (*domain.AggregatedResponse, error) {
	since := time.Now().Add(-period)

	return s.repo.QueryHighestSinceByExchange(exchange, symbol, since)
}
