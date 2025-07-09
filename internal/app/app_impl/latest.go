package app_impl

import (
	"errors"
	"marketflow/internal/domain"
	"strings"
)

func (s *APIService) GetAggregatedPriceForSymbol(symbol string) (*domain.AggregatedResponse, error) {
	symbol = strings.ToUpper(symbol)

	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}

	data, err := s.repo.GetPriceForSymbol(symbol)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *APIService) GetAggregatedPriceForExchange(path string) (*domain.AggregatedResponse, error) {
	parts := strings.Split(path, "/")
	exchange := parts[0]
	symbol := parts[1]
	symbol = strings.ToUpper(symbol)
	data, err := s.repo.GetPriceForExchange(exchange, symbol)
	if err != nil {
		return nil, err
	}

	return data, err
}
