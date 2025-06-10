package binance

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
			conn, err := connectAndHandshake(u, "stream.binance.com")
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
			opcode, frame, err := readFrameBinance(b.conn)
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
				return
			}
			fmt.Println("[binance] ping sent")
		}
	}
}

// --- Вспомогательные функции ---

func connectAndHandshake(u url.URL, host string) (*tls.Conn, error) {
	conf := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", u.Host, conf)
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
		return nil, fmt.Errorf("handshake write failed: %w", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("handshake read failed: %w", err)
	}
	if !bytes.Contains(buf[:n], []byte("101 Switching Protocols")) {
		return nil, fmt.Errorf("unexpected handshake response: %s", buf[:n])
	}

	return conn, nil
}

func readFrameBinance(r io.Reader) (opcode byte, payload []byte, err error) {
	header := make([]byte, 2)
	if _, err = io.ReadFull(r, header); err != nil {
		return 0, nil, fmt.Errorf("read header: %w", err)
	}

	fin := header[0]&0x80 != 0
	opcode = header[0] & 0x0F
	masked := header[1]&0x80 != 0
	payloadLen := int(header[1] & 0x7F)

	if !fin {
		return 0, nil, fmt.Errorf("fragmented frames not supported")
	}

	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		if _, err = io.ReadFull(r, ext); err != nil {
			return 0, nil, fmt.Errorf("read extended payload (126): %w", err)
		}
		payloadLen = int(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err = io.ReadFull(r, ext); err != nil {
			return 0, nil, fmt.Errorf("read extended payload (127): %w", err)
		}
		payloadLen = int(binary.BigEndian.Uint64(ext))
	}

	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		if _, err = io.ReadFull(r, maskKey); err != nil {
			return 0, nil, fmt.Errorf("read mask key: %w", err)
		}
	}

	payload = make([]byte, payloadLen)
	if _, err = io.ReadFull(r, payload); err != nil {
		return 0, nil, fmt.Errorf("read payload: %w", err)
	}

	if masked {
		for i := 0; i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	switch opcode {
	case 0x1: // text
		return opcode, payload, nil
	case 0x8: // close
		return opcode, nil, io.EOF
	case 0x9, 0xA: // ping/pong
		return opcode, nil, nil
	default:
		return opcode, nil, fmt.Errorf("unsupported opcode: %d", opcode)
	}
}

func writePingFrame(conn net.Conn) error {
	frame := []byte{0x89, 0x00}
	_, err := conn.Write(frame)
	if err != nil {
		return fmt.Errorf("WritePingFrame error: %w", err)
	}
	return nil
}

func generateWebSocketKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
