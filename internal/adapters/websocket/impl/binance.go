// internal/adapters/websocket/impl/binance_adapter.go
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
	binanceHost = "stream.binance.com:443"
	binancePath = "/ws/btcusdt@trade"
)

type BinanceAdapter struct {
	conn   *tls.Conn
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewBinanceAdapter() *BinanceAdapter {
	return &BinanceAdapter{
		stopCh: make(chan struct{}),
	}
}

// Start устанавливает соединение и начинает читать данные, отправляя их в out канал.
func (b *BinanceAdapter) Start(out chan<- domain.MarketData) error {
	u := url.URL{Scheme: "wss", Host: binanceHost, Path: binancePath}
	fmt.Println("Connecting to", u.String())

	conn, err := tls.Dial("tcp", binanceHost, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	b.conn = conn

	req := bytes.NewBufferString(fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: 123456==\r\n\r\n",
		binancePath, "stream.binance.com",
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

				var msg struct {
					Price     string `json:"p"`
					Quantity  string `json:"q"`
					EventTime int64  `json:"E"`
					Symbol    string `json:"s"`
				}

				if err := json.Unmarshal(frame, &msg); err != nil {
					continue
				}

				price, _ := helpers.ParseFloat(msg.Price)
				volume, _ := helpers.ParseFloat(msg.Quantity)

				data := domain.MarketData{
					Exchange: "binance",
					Symbol:   msg.Symbol,
					Price:    price,
					Volume:   volume,
					Time:     msg.EventTime,
				}

				out <- data
			}
		}
	}()

	return nil
}

// Stop закрывает соединение и ждёт завершения горутины.
func (b *BinanceAdapter) Stop() error {
	close(b.stopCh)
	b.wg.Wait()
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
