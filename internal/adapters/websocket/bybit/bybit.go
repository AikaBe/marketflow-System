package bybit

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"marketflow/internal/domain"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	bybitHost = "stream.bybit.com"
	bybitPath = "/v5/trade"
)

type BybitAdapter struct {
	conn       *tls.Conn
	stopCh     chan struct{}
	connClosed chan struct{}
	wg         sync.WaitGroup
}

func NewBybitAdapter() *BybitAdapter {
	return &BybitAdapter{
		stopCh:     make(chan struct{}),
		connClosed: make(chan struct{}),
	}
}

func (b *BybitAdapter) Start(out chan<- domain.MarketData) error {
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()

		for {
			select {
			case <-b.stopCh:
				return
			default:
			}

			fmt.Println("[bybit] connecting...")
			u := url.URL{Scheme: "wss", Host: bybitHost, Path: bybitPath}
			conn, err := connectAndHandshake(u, bybitHost)
			if err != nil {
				fmt.Println("[bybit] connection error:", err)
				time.Sleep(5 * time.Second)
				continue
			}
			b.conn = conn

			subscribeMsg := []byte(`{"op": "subscribe", "args": ["orderbook.25.SOLUSDT_SOL/USDT"]}`)
			_, err = conn.Write(subscribeMsg)
			if err != nil {
				fmt.Println("[bybit] subscribe error:", err)
				conn.Close()
				time.Sleep(5 * time.Second)
				continue
			}

			// Закрываем и пересоздаем канал на случай переподключения
			b.connClosed = make(chan struct{})

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
	close(b.stopCh)
	b.wg.Wait()
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *BybitAdapter) readLoop(out chan<- domain.MarketData) {
	defer close(b.connClosed)

	for {
		select {
		case <-b.stopCh:
			return
		default:
		}

		opcode, frame, err := readFrameBybit(b.conn)
		if err != nil {
			fmt.Println("[bybit] read error:", err)
			return
		}
		if opcode == 0xA {
			fmt.Println("[bybit] pong received")
			continue
		}
		if opcode != 0x1 {
			continue
		}

		var msg struct {
			Topic string `json:"topic"`
			Data  []struct {
				Price     string `json:"price"`
				Quantity  string `json:"size"`
				Timestamp string `json:"timestamp"`
			} `json:"data"`
		}
		if err := json.Unmarshal(frame, &msg); err != nil {
			fmt.Println("[bybit] unmarshal error:", err)
			continue
		}

		if len(msg.Data) == 0 {
			continue
		}

		price, err := strconv.ParseFloat(msg.Data[0].Price, 64)
		if err != nil {
			fmt.Println("[bybit] price parse error:", err)
			continue
		}
		volume, err := strconv.ParseFloat(msg.Data[0].Quantity, 64)
		if err != nil {
			fmt.Println("[bybit] volume parse error:", err)
			continue
		}

		timestamp, err := time.Parse(time.RFC3339, msg.Data[0].Timestamp)
		if err != nil {
			fmt.Println("[bybit] timestamp parse error:", err)
			continue
		}

		out <- domain.MarketData{
			Exchange: "bybit",
			Symbol:   "BTCUSDT",
			Price:    price,
			Volume:   volume,
			Time:     timestamp.Unix(),
		}
	}
}

func (b *BybitAdapter) pingLoop() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-b.connClosed:
			return
		case <-ticker.C:
			if b.conn == nil {
				return
			}
			err := writePingFrame(b.conn)
			if err != nil {
				fmt.Println("[bybit] ping write error:", err)
				return
			}
			fmt.Println("[bybit] ping sent")
		}
	}
}

// --- вспомогательные функции ---

func connectAndHandshake(u url.URL, host string) (*tls.Conn, error) {
	conf := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", u.Host+":443", conf)
	if err != nil {
		return nil, fmt.Errorf("TLS dial failed: %w", err)
	}

	secKey := generateWebSocketKey()
	req := fmt.Sprintf(
		"GET %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Version: 13\r\n"+
			"Sec-WebSocket-Key: %s\r\n\r\n",
		u.Path, host, secKey,
	)

	_, err = conn.Write([]byte(req))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake write failed: %w", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake read failed: %w", err)
	}
	if !bytes.Contains(buf[:n], []byte("101 Switching Protocols")) {
		conn.Close()
		return nil, fmt.Errorf("unexpected handshake response: %s", buf[:n])
	}

	return conn, nil
}

func readFrameBybit(r io.Reader) (byte, []byte, error) {
	header := make([]byte, 2)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return 0, nil, fmt.Errorf("read header: %w", err)
	}

	fin := header[0]&0x80 != 0
	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	payloadLen := int(header[1] & 0x7F)

	if !fin {
		return 0, nil, fmt.Errorf("fragmented frames not supported")
	}

	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		_, err = io.ReadFull(r, ext)
		if err != nil {
			return 0, nil, fmt.Errorf("read extended payload (126): %w", err)
		}
		payloadLen = int(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		_, err = io.ReadFull(r, ext)
		if err != nil {
			return 0, nil, fmt.Errorf("read extended payload (127): %w", err)
		}
		payloadLen = int(binary.BigEndian.Uint64(ext))
	}

	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		_, err = io.ReadFull(r, maskKey)
		if err != nil {
			return 0, nil, fmt.Errorf("read mask key: %w", err)
		}
	}

	payload := make([]byte, payloadLen)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return 0, nil, fmt.Errorf("read payload: %w", err)
	}

	if masked {
		for i := 0; i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	switch opcode {
	case 0x1, 0x8, 0x9, 0xA:
		return opcode, payload, nil
	default:
		return opcode, nil, fmt.Errorf("unsupported opcode: %d", opcode)
	}
}

func writePingFrame(conn net.Conn) error {
	frame := []byte{0x89, 0x00}
	_, err := conn.Write(frame)
	if err != nil {
		return fmt.Errorf("writePingFrame error: %w", err)
	}
	return nil
}

func generateWebSocketKey() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}
