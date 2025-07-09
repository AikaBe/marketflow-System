package api

import (
	"errors"
	"log/slog"
	"marketflow/internal/domain"
	"strings"
	"time"
)

func (s *APIService) GetHighestBySymbol(symbol string) (*domain.AggregatedResponse, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	if symbol == "" {
		slog.Warn("GetHighestBySymbol: empty symbol")
		return nil, errors.New("symbol cannot be empty")
	}

	slog.Info("GetHighestBySymbol called", "symbol", symbol)
	data, err := s.repo.GetHighestBySymbol(symbol)
	if err != nil {
		slog.Error("GetHighestBySymbol failed", "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("GetHighestBySymbol: no data", "symbol", symbol)
		return nil, errors.New("no data found for symbol: " + symbol)
	}

	slog.Info("GetHighestBySymbol success", "symbol", symbol, "max", data.Max)
	return data, nil
}

func (s *APIService) GetHighestByExchange(path string) (*domain.AggregatedResponse, error) {
	slog.Info("GetHighestByExchange called", "path", path)

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		slog.Warn("GetHighestByExchange: invalid path format", "path", path)
		return nil, errors.New("invalid path format: expected /exchange/symbol")
	}

	exchange := strings.TrimSpace(parts[0])
	symbol := strings.ToUpper(strings.TrimSpace(parts[1]))

	if exchange == "" || symbol == "" {
		slog.Warn("GetHighestByExchange: exchange or symbol is empty", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("exchange and symbol must not be empty")
	}

	data, err := s.repo.GetHighestByExchange(exchange, symbol)
	if err != nil {
		slog.Error("GetHighestByExchange failed", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("GetHighestByExchange: no data", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("no data found for exchange/symbol: " + exchange + "/" + symbol)
	}

	slog.Info("GetHighestByExchange success", "exchange", exchange, "symbol", symbol, "max", data.Max)
	return data, nil
}

func (s *APIService) GetHighestByPeriod(symbol string, period time.Duration) (*domain.AggregatedResponse, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		slog.Warn("GetHighestByPeriod: empty symbol")
		return nil, errors.New("symbol cannot be empty")
	}

	since := time.Now().Add(-period)
	slog.Info("GetHighestByPeriod called", "symbol", symbol, "since", since)

	data, err := s.repo.QueryHighestPriceSince(symbol, since)
	if err != nil {
		slog.Error("GetHighestByPeriod failed", "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("GetHighestByPeriod: no data", "symbol", symbol)
		return nil, errors.New("no data found for symbol in period")
	}

	slog.Info("GetHighestByPeriod success", "symbol", symbol, "max", data.Max)
	return data, nil
}

func (s *APIService) QueryHighestSinceByExchange(exchange, symbol string, period time.Duration) (*domain.AggregatedResponse, error) {
	exchange = strings.TrimSpace(exchange)
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	if exchange == "" || symbol == "" {
		slog.Warn("QueryHighestSinceByExchange: empty exchange or symbol", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("exchange and symbol must not be empty")
	}

	since := time.Now().Add(-period)
	slog.Info("QueryHighestSinceByExchange called", "exchange", exchange, "symbol", symbol, "since", since)

	data, err := s.repo.QueryHighestSinceByExchange(exchange, symbol, since)
	if err != nil {
		slog.Error("QueryHighestSinceByExchange failed", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("QueryHighestSinceByExchange: no data", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("no data found for exchange/symbol in period")
	}

	slog.Info("QueryHighestSinceByExchange success", "exchange", exchange, "symbol", symbol, "max", data.Max)
	return data, nil
}
