package mode

import (
	"context"
	"log"
	"marketflow/internal/adapters/generator"
	"marketflow/internal/adapters/websocket"
	"marketflow/internal/domain"
	"sync"
)

type Mode int

const (
	ModeLive Mode = iota
	ModeTest
)

type Manager struct {
	current Mode
	cancel  context.CancelFunc
	mu      sync.Mutex
	out     chan<- domain.PriceUpdate
}

func NewModeManager(out chan<- domain.PriceUpdate) *Manager {
	return &Manager{
		current: ModeLive,
		out:     out,
	}
}

func (m *Manager) SetMode(ctx context.Context, mode Mode) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		log.Println("[MODE] Cancelling previous mode")
		m.cancel()
	}

	newCtx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.current = mode

	switch mode {
	case ModeLive:
		log.Println("[MODE] Switched to Live Mode")
		go websocket.StartReaders(m.out)
	case ModeTest:
		log.Println("[MODE] Switched to Test Mode")
		go generator.StartTestGenerators(newCtx, m.out)
	}
}

func (m *Manager) GetMode() Mode {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.current
}
