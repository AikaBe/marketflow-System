package impl

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"marketflow/internal/adapters/websocket/impl/helpers"
	"marketflow/internal/domain"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	binanceHost = "stream.binance.com:9443"
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

func (b *BinanceAdapter) Start(out chan<- domain.MarketData) error {
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()

		for {
			select {
			case <-b.stopCh:
				return
			default:
			}

			fmt.Println("[binance] connecting...")
			u := url.URL{Scheme: "wss", Host: binanceHost, Path: binancePath}
			conn, err := helpers.ConnectAndHandshake(u, "stream.binance.com")
			if err != nil {
				fmt.Println("[binance] connection error:", err)
				time.Sleep(5 * time.Second)
				continue
			}
			b.conn = conn

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

			// Ждём, пока read или ping завершится (в случае ошибки)
			innerWg.Wait()
			fmt.Println("[binance] reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func (b *BinanceAdapter) Stop() error {
	close(b.stopCh)
	b.wg.Wait()
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *BinanceAdapter) readLoop(out chan<- domain.MarketData) {
	for {
		select {
		case <-b.stopCh:
			return
		default:
			opcode, frame, err := helpers.ReadFrame(b.conn)
			if err != nil {
				fmt.Println("[binance] read error:", err)
				return
			}
			if opcode == 0xA {
				fmt.Println("[binance] pong received")
				continue
			}
			if opcode != 0x1 {
				continue
			}

			var msg struct {
				Price     string `json:"p"`
				Quantity  string `json:"q"`
				Timestamp int64  `json:"T"`
			}
			if err := json.Unmarshal(frame, &msg); err != nil {
				fmt.Println("[binance] unmarshal error:", err)
				continue
			}

			price, err := strconv.ParseFloat(msg.Price, 64)
			if err != nil {
				fmt.Println("[binance] price parse error:", err)
				continue
			}
			volume, err := strconv.ParseFloat(msg.Quantity, 64)
			if err != nil {
				fmt.Println("[binance] volume parse error:", err)
				continue
			}

			out <- domain.MarketData{
				Exchange: "binance",
				Symbol:   "BTCUSDT",
				Price:    price,
				Volume:   volume,
				Time:     msg.Timestamp,
			}
		}
	}
}

func (b *BinanceAdapter) pingLoop() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ticker.C:
			err := writePingFrame(b.conn)
			if err != nil {
				fmt.Println("[binance] ping write error:", err)
				return // это завершит pingLoop и вызовет reconnect
			}
			fmt.Println("[binance] ping sent")
		}
	}
}

func writePingFrame(conn net.Conn) error {
	// Frame формат: fin=1, opcode=0x9 (ping), no mask, no payload
	frame := []byte{0x89, 0x00}
	_, err := conn.Write(frame)
	if err != nil {
		return fmt.Errorf("WritePingFrame error: %w", err)
	}
	return nil
}
