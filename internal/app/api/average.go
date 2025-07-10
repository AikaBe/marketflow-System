package api

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"marketflow/internal/domain"
)

func (s *APIService) GetAvgBySymbol(symbol string) (*domain.AggregatedResponse, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		slog.Warn("GetAvgBySymbol: symbol is empty")
		return nil, errors.New("symbol cannot be empty")
	}

	slog.Info("GetAvgBySymbol called", "symbol", symbol)
	data, err := s.repo.GetAvgBySymbol(symbol)
	if err != nil {
		slog.Error("GetAvgBySymbol failed", "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("GetAvgBySymbol: no data found", "symbol", symbol)
		return nil, errors.New("no data found for symbol: " + symbol)
	}

	slog.Info("GetAvgBySymbol success", "symbol", symbol, "avg", data.Avg)
	return data, nil
}

func (s *APIService) GetAvgByExchange(path string) (*domain.AggregatedResponse, error) {
	slog.Info("GetAvgByExchange called", "path", path)

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		slog.Warn("GetAvgByExchange: invalid path", "path", path)
		return nil, errors.New("invalid path format: expected /exchange/symbol")
	}

	exchange := strings.TrimSpace(parts[0])
	symbol := strings.ToUpper(strings.TrimSpace(parts[1]))

	if exchange == "" || symbol == "" {
		slog.Warn("GetAvgByExchange: exchange or symbol is empty", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("exchange and symbol must not be empty")
	}

	data, err := s.repo.GetAvgByExchange(exchange, symbol)
	if err != nil {
		slog.Error("GetAvgByExchange failed", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("GetAvgByExchange: no data found", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("no data found for exchange/symbol: " + exchange + "/" + symbol)
	}

	slog.Info("GetAvgByExchange success", "exchange", exchange, "symbol", symbol, "avg", data.Avg)
	return data, nil
}

func (s *APIService) QueryAvgSinceByExchange(exchange, symbol string, period time.Duration) (*domain.AggregatedResponse, error) {
	slog.Info("QueryAvgSinceByExchange called", "exchange", exchange, "symbol", symbol, "period", period)

	exchange = strings.TrimSpace(exchange)
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	if exchange == "" || symbol == "" {
		slog.Warn("QueryAvgSinceByExchange: exchange or symbol is empty")
		return nil, errors.New("exchange and symbol must not be empty")
	}

	since := time.Now().Add(-period)
	data, err := s.repo.QueryAvgSinceByExchange(exchange, symbol, since)
	if err != nil {
		slog.Error("QueryAvgSinceByExchange failed", "exchange", exchange, "symbol", symbol, "err", err)
		return nil, err
	}
	if data == nil {
		slog.Warn("QueryAvgSinceByExchange: no data found", "exchange", exchange, "symbol", symbol)
		return nil, errors.New("no data found for exchange/symbol in given period")
	}

	slog.Info("QueryAvgSinceByExchange success", "exchange", exchange, "symbol", symbol, "avg", data.Avg)
	return data, nil
}
