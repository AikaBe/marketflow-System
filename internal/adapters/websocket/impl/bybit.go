package impl

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"marketflow/internal/adapters/websocket/impl/helpers"
	"marketflow/internal/domain"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	bybitHost = "stream.bybit.com:443"
	bybitPath = "/v5/public/linear"
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
	fmt.Println("[bybit] starting adapter...")

	b.wg.Add(1)

	go func() {
		defer b.wg.Done()

		for {
			select {
			case <-b.stopCh:
				fmt.Println("[bybit] stop signal received, exiting main loop")
				return
			default:
			}

			fmt.Println("[bybit] connecting...")
			u := url.URL{Scheme: "wss", Host: bybitHost, Path: bybitPath}
			conn, err := helpers.ConnectAndHandshake(u, "stream.bybit.com")
			if err != nil {
				fmt.Println("[bybit] connection error:", err)
				time.Sleep(5 * time.Second)
				continue
			}
			b.conn = conn

			subscribe := map[string]interface{}{
				"op":   "subscribe",
				"args": []string{"trade.BTCUSDT"},
			}
			msgJSON, _ := json.Marshal(subscribe)

			if err := helpers.WriteFrame(conn, msgJSON); err != nil {
				fmt.Println("[bybit] subscription failed:", err)
				conn.Close()
				time.Sleep(5 * time.Second)
				continue
			}
			fmt.Println("[bybit] subscribed to trade.BTCUSDT")

			var innerWg sync.WaitGroup
			innerWg.Add(2)

			go func() {
				defer innerWg.Done()
				b.readLoop(out)
			}()

			go func() {
				defer innerWg.Done()
				b.pingLoop()
			}()

			innerWg.Wait()
			fmt.Println("[bybit] reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func (b *BybitAdapter) Stop() error {
	fmt.Println("[bybit] stopping adapter...")
	close(b.stopCh)
	b.wg.Wait()
	if b.conn != nil {
		err := b.conn.Close()
		if err != nil {
			fmt.Println("[bybit] connection close error:", err)
			return err
		}
		fmt.Println("[bybit] connection closed successfully")
	}
	fmt.Println("[bybit] adapter stopped")
	return nil
}

func (b *BybitAdapter) readLoop(out chan<- domain.MarketData) {
	fmt.Println("[bybit] starting read loop")
	for {
		select {
		case <-b.stopCh:
			fmt.Println("[bybit] read loop received stop signal")
			return
		default:
			opcode, frame, err := helpers.ReadFrame(b.conn)
			if err != nil {
				if err == io.EOF {
					fmt.Println("[bybit] server closed the connection (EOF)")
				} else {
					fmt.Println("[bybit] read error:", err)
				}
				return
			}

			if opcode == 0xA {
				fmt.Println("[bybit] pong received")
				continue
			}
			if opcode != 0x1 {
				fmt.Printf("[bybit] skipping non-text frame (opcode: %d)\n", opcode)
				continue
			}

			var msg struct {
				Topic string `json:"topic"`
				Data  []struct {
					Symbol    string `json:"symbol"`
					Price     string `json:"price"`
					Size      string `json:"size"`
					Timestamp int64  `json:"timestamp"`
				} `json:"data"`
			}

			if err := json.Unmarshal(frame, &msg); err != nil {
				fmt.Println("[bybit] unmarshal error:", err)
				continue
			}
			if msg.Topic != "trade.BTCUSDT" {
				fmt.Printf("[bybit] received irrelevant topic: %s\n", msg.Topic)
				continue
			}

			for _, trade := range msg.Data {
				price, err := strconv.ParseFloat(trade.Price, 64)
				if err != nil {
					fmt.Println("[bybit] price parse error:", err)
					continue
				}
				volume, err := strconv.ParseFloat(trade.Size, 64)
				if err != nil {
					fmt.Println("[bybit] volume parse error:", err)
					continue
				}

				md := domain.MarketData{
					Exchange: "bybit",
					Symbol:   trade.Symbol,
					Price:    price,
					Volume:   volume,
					Time:     trade.Timestamp,
				}

				select {
				case out <- md:
					fmt.Printf("[bybit] emitted trade: %+v\n", md)
				case <-b.stopCh:
					fmt.Println("[bybit] read loop stopped while sending")
					return
				}
			}
		}
	}
}

func (b *BybitAdapter) pingLoop() {
	fmt.Println("[bybit] starting ping loop")
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			fmt.Println("[bybit] ping loop received stop signal")
			return
		case <-ticker.C:
			err := helpers.WritePingFrame(b.conn)
			if err != nil {
				fmt.Println("[bybit] ping write error:", err)
				return // выход приведёт к реконнекту
			}
			fmt.Println("[bybit] ping sent")
		}
	}
}
