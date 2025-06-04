package usecase

import (
	"context"
	"marketflow/internal/adapters/websocket"
	"marketflow/internal/domain"
	"sync"
)

type MarketDataService struct {
	providers []websocket.MarketDataProvider
	out       chan domain.MarketData
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewMarketDataService(providers ...websocket.MarketDataProvider) *MarketDataService {
	return &MarketDataService{
		providers: providers,
		out:       make(chan domain.MarketData),
	}
}

func (m *MarketDataService) Start(ctx context.Context) <-chan domain.MarketData {
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	for _, provider := range m.providers {
		m.wg.Add(1)
		go func(p websocket.MarketDataProvider) {
			defer m.wg.Done()
			p.Start(m.out)
		}(provider)
	}

	return m.out
}

func (m *MarketDataService) Stop() {
	for _, p := range m.providers {
		_ = p.Stop()
	}
	m.cancel()
	m.wg.Wait()
	close(m.out)
}
