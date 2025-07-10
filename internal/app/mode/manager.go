package mode

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"marketflow/internal/adapters/generator"
	"marketflow/internal/adapters/websocket"
	"marketflow/internal/domain"
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

func (m *Manager) SetMode(ctx context.Context, mode Mode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	slog.Info("SetMode called", "requested_mode", mode)

	if mode != ModeLive && mode != ModeTest {
		slog.Error("Invalid mode value", "mode", mode)
		return errors.New("invalid mode")
	}

	if m.cancel != nil {
		slog.Info("Cancelling previous mode", "previous_mode", m.current)
		m.cancel()
	}

	newCtx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.current = mode

	switch mode {
	case ModeLive:
		slog.Info("Switched to Live Mode")
		go websocket.StartReaders(m.out)
	case ModeTest:
		slog.Info("Switched to Test Mode")
		go generator.StartTestGenerators(newCtx, m.out)
	}

	return nil
}

func (m *Manager) GetMode() Mode {
	m.mu.Lock()
	defer m.mu.Unlock()
	slog.Info("GetMode called", "current_mode", m.current)
	return m.current
}
