package api

import (
	"errors"
	"log/slog"
	"marketflow/internal/domain"
	"strings"
	"time"
)

func (s *APIService) GetLowestBySymbol(symbol string) (*domain.AggregatedResponse, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	if symbol == "" {
		slog.Warn("GetLowestBySymbol: symbol is empty")
		return nil, errors.New("symbol cannot be empty")
	}

	slog.Info("GetLowestBySymbol called", "symbol", symbol)
	data, err := s.repo.GetLowestBySymbol(symbol)
	if err != nil {
		slog.Error("GetLowestBySymbol failed", "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("No lowest price data found", "symbol", symbol)
		return nil, errors.New("no data found for symbol: " + symbol)
	}

	slog.Info("GetLowestBySymbol success", "symbol", symbol, "min", data.Min)
	return data, nil
}

func (s *APIService) GetLowestByExchange(path string) (*domain.AggregatedResponse, error) {
	slog.Info("GetLowestByExchange called", "path", path)

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

	data, err := s.repo.GetLowestByExchange(exchange, symbol)
	if err != nil {
		slog.Error("GetLowestByExchange failed", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("No lowest price data found", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("no data found for exchange/symbol: " + exchange + "/" + symbol)
	}

	slog.Info("GetLowestByExchange success", "exchange", exchange, "symbol", symbol, "min", data.Min)
	return data, nil
}

func (s *APIService) GetLowestByPeriod(symbol string, period time.Duration) (*domain.AggregatedResponse, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		slog.Warn("GetLowestByPeriod: symbol is empty")
		return nil, errors.New("symbol cannot be empty")
	}

	since := time.Now().Add(-period)
	slog.Info("GetLowestByPeriod called", "symbol", symbol, "since", since)

	data, err := s.repo.QueryLowestPriceSince(symbol, since)
	if err != nil {
		slog.Error("QueryLowestPriceSince failed", "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("No lowest price data found for period", "symbol", symbol, "since", since)
		return nil, errors.New("no data found for symbol: " + symbol + " in period")
	}

	slog.Info("GetLowestByPeriod success", "symbol", symbol, "min", data.Min)
	return data, nil
}

func (s *APIService) QueryLowestSinceByExchange(exchange, symbol string, period time.Duration) (*domain.AggregatedResponse, error) {
	exchange = strings.TrimSpace(exchange)
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if exchange == "" || symbol == "" {
		slog.Warn("QueryLowestSinceByExchange: empty values", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("exchange and symbol must not be empty")
	}

	since := time.Now().Add(-period)
	slog.Info("QueryLowestSinceByExchange called", "exchange", exchange, "symbol", symbol, "since", since)

	data, err := s.repo.QueryLowestSinceByExchange(exchange, symbol, since)
	if err != nil {
		slog.Error("QueryLowestSinceByExchange failed", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("No lowest price data found for exchange and period", "exchange", exchange, "symbol", symbol, "since", since)
		return nil, errors.New("no data found for exchange/symbol: " + exchange + "/" + symbol + " in period")
	}

	slog.Info("QueryLowestSinceByExchange success", "exchange", exchange, "symbol", symbol, "min", data.Min)
	return data, nil
}
