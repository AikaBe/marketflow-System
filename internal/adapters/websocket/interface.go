// internal/adapters/websocket/interface.go
package websocket

import "marketflow/internal/domain"

type MarketDataProvider interface {
	Start(out chan<- domain.MarketData) error
	Stop() error
}
