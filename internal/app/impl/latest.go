package impl

import (
	"errors"
	"marketflow/internal/adapters/postgres"
	"marketflow/internal/domain"
	"strings"
)

type LatestService struct {
	Repo *postgres.Adapter
}

func (s *LatestService) GetAggregatedPriceForSymbol(symbol string) (*domain.AggregatedResponse, error) {
	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}

	data, err := s.Repo.GetAggregatedPriceForSymbol(symbol)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *LatestService) GetAggregatedPriceForExchange(path string) (*domain.AggregatedResponse, error) {
	parts := strings.Split(path, "/")
	symbol := parts[len(parts)-1]
	exchange := parts[len(parts)-2]
	data, err := s.Repo.GetAggregatedPriceForExchange(exchange, symbol)
	if err != nil {
		return nil, err
	}

	return data, err
}
