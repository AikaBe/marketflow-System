package impl

import (
	"errors"
	"marketflow/internal/adapters/postgres"
	"marketflow/internal/domain"
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
