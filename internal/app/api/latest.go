package api

import (
	"errors"
	"log/slog"
	"strings"

	"marketflow/internal/domain"
)

func (s *APIService) GetAggregatedPriceForSymbol(symbol string) (*domain.AggregatedResponse, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	if symbol == "" {
		slog.Warn("GetAggregatedPriceForSymbol: symbol is empty")
		return nil, errors.New("symbol cannot be empty")
	}

	slog.Info("GetAggregatedPriceForSymbol called", "symbol", symbol)
	data, err := s.repo.GetPriceForSymbol(symbol)
	if err != nil {
		slog.Error("GetPriceForSymbol failed", "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("GetPriceForSymbol: no data found", "symbol", symbol)
		return nil, errors.New("no data found for symbol: " + symbol)
	}

	slog.Info("GetPriceForSymbol success", "symbol", symbol, "avg", data.Avg)
	return data, nil
}

func (s *APIService) GetAggregatedPriceForExchange(path string) (*domain.AggregatedResponse, error) {
	slog.Info("GetAggregatedPriceForExchange called", "path", path)

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		slog.Warn("Invalid path format", "path", path)
		return nil, errors.New("invalid path format, expected /exchange/symbol")
	}

	exchange := strings.TrimSpace(parts[0])
	symbol := strings.ToUpper(strings.TrimSpace(parts[1]))

	if exchange == "" || symbol == "" {
		slog.Warn("Exchange or symbol is empty", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("exchange and symbol must not be empty")
	}

	data, err := s.repo.GetPriceForExchange(exchange, symbol)
	if err != nil {
		slog.Error("GetPriceForExchange failed", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("GetPriceForExchange: no data", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("no data found for exchange/symbol: " + exchange + "/" + symbol)
	}

	slog.Info("GetPriceForExchange success", "exchange", exchange, "symbol", symbol, "avg", data.Avg)
	return data, nil
}
