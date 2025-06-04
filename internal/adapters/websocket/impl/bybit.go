// internal/adapters/websocket/impl/bybit_adapter.go
package impl

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"marketflow/internal/adapters/websocket/impl/helpers"
	"marketflow/internal/domain"
	"net/url"
	"sync"
)

const (
	bybitHost = "stream.bybit.com:443"
	bybitPath = "/v5/public/quote/ws"
)

type BybitAdapter struct {
	conn   *tls.Conn
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewBybitAdapter() *BybitAdapter {
	return &BybitAdapter{
		stopCh: make(chan struct{}),
	}
}

func (b *BybitAdapter) Start(out chan<- domain.MarketData) error {
	u := url.URL{Scheme: "wss", Host: bybitHost, Path: bybitPath}
	fmt.Println("Connecting to", u.String())

	conn, err := tls.Dial("tcp", bybitHost, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	b.conn = conn

	// Рукопожатие WebSocket (вручную)
	req := bytes.NewBufferString(fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: 123456==\r\n\r\n",
		bybitPath, "stream.bybit.com",
	))

	_, err = conn.Write(req.Bytes())
	if err != nil {
		return fmt.Errorf("handshake write error: %w", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("handshake read error: %w", err)
	}
	if !bytes.Contains(buf[:n], []byte("101 Switching Protocols")) {
		return fmt.Errorf("unexpected handshake response:\n%s", buf[:n])
	}

	// Подписка на канал сделок BTCUSDT
	subscribeMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": []string{"trade.BTCUSDT"},
	}
	msgJSON, _ := json.Marshal(subscribeMsg)
	err = helpers.WriteFrame(conn, msgJSON)
	if err != nil {
		return fmt.Errorf("subscribe write error: %w", err)
	}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		defer conn.Close()

		for {
			select {
			case <-b.stopCh:
				return
			default:
				frame, err := helpers.ReadFrame(conn)
				if err != nil {
					fmt.Println("read frame error:", err)
					return
				}

				var rawMsg struct {
					Topic string `json:"topic"`
					Data  []struct {
						Symbol    string  `json:"symbol"`
						Price     float64 `json:"price,string"`
						Size      float64 `json:"size,string"`
						Timestamp int64   `json:"timestamp"`
					} `json:"data"`
				}

				err = json.Unmarshal(frame, &rawMsg)
				if err != nil {
					continue
				}
				if rawMsg.Topic != "trade.BTCUSDT" {
					continue
				}

				for _, trade := range rawMsg.Data {
					data := domain.MarketData{
						Exchange: "bybit",
						Symbol:   trade.Symbol,
						Price:    trade.Price,
						Volume:   trade.Size,
						Time:     trade.Timestamp,
					}
					out <- data
				}
			}
		}
	}()

	return nil
}

func (b *BybitAdapter) Stop() error {
	close(b.stopCh)
	b.wg.Wait()
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
