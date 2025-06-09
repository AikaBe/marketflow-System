package impl

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"marketflow/internal/domain"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
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
	mu     sync.Mutex
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

			conn, err := b.connectWithWebSocketHandshake()
			if err != nil {
				fmt.Println("[bybit] connection error:", err)
				time.Sleep(5 * time.Second)
				continue
			}

			b.setConn(conn)

			if err := b.subscribe(); err != nil {
				fmt.Println("[bybit] subscription failed:", err)
				b.closeConn()
				time.Sleep(5 * time.Second)
				continue
			}

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

			b.closeConn()
			fmt.Println("[bybit] reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func (b *BybitAdapter) setConn(conn *tls.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.conn = conn
}

func (b *BybitAdapter) getConn() *tls.Conn {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.conn
}

func (b *BybitAdapter) closeConn() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn != nil {
		_ = b.conn.Close()
		b.conn = nil
	}
}

// connectWithWebSocketHandshake выполняет низкоуровневое подключение и ручной WS-хендшейк
func (b *BybitAdapter) connectWithWebSocketHandshake() (*tls.Conn, error) {
	u := url.URL{Scheme: "wss", Host: bybitHost, Path: bybitPath}
	fmt.Println("[bybit] connecting to", u.String())

	conn, err := tls.Dial("tcp", bybitHost, nil)
	if err != nil {
		return nil, err
	}

	// Генерируем Sec-WebSocket-Key
	key := make([]byte, 16)
	if _, err = rand.Read(key); err != nil {
		conn.Close()
		return nil, err
	}
	secWebSocketKey := base64.StdEncoding.EncodeToString(key)

	// Формируем запрос хендшейка
	req := fmt.Sprintf("GET %s HTTP/1.1\r\n", bybitPath) +
		fmt.Sprintf("Host: %s\r\n", bybitHost) +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		fmt.Sprintf("Sec-WebSocket-Key: %s\r\n", secWebSocketKey) +
		"Sec-WebSocket-Version: 13\r\n" +
		"\r\n"

	if _, err = conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)

	statusLine, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !strings.Contains(statusLine, "101") {
		conn.Close()
		return nil, fmt.Errorf("websocket handshake failed, status: %s", strings.TrimSpace(statusLine))
	}

	// Читаем заголовки
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // конец заголовков
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if !strings.EqualFold(headers["Upgrade"], "websocket") || !strings.Contains(strings.ToLower(headers["Connection"]), "upgrade") {
		conn.Close()
		return nil, errors.New("invalid websocket upgrade headers")
	}

	// Опционально можно добавить проверку Sec-WebSocket-Accept

	return conn, nil
}

func (b *BybitAdapter) subscribe() error {
	subscribe := map[string]interface{}{
		"op":   "subscribe",
		"args": []string{"trade.BTCUSDT"},
	}
	msgJSON, err := json.Marshal(subscribe)
	if err != nil {
		return err
	}

	conn := b.getConn()
	if conn == nil {
		return errors.New("connection is nil")
	}

	return writeFrame(conn, msgJSON)
}

func (b *BybitAdapter) Stop() error {
	fmt.Println("[bybit] stopping adapter...")
	// Закрываем канал остановки, если он ещё открыт
	select {
	case <-b.stopCh:
		// уже закрыт
	default:
		close(b.stopCh)
	}
	b.wg.Wait()
	b.closeConn()
	fmt.Println("[bybit] adapter stopped")
	return nil
}

func (b *BybitAdapter) readLoop(out chan<- domain.MarketData) {
	fmt.Println("[bybit] starting read loop")

	conn := b.getConn()
	if conn == nil {
		fmt.Println("[bybit] readLoop: connection is nil")
		return
	}

	for {
		select {
		case <-b.stopCh:
			return
		default:
		}

		opcode, frame, err := readFrame(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("[bybit] server closed the connection (EOF)")
			} else {
				fmt.Println("[bybit] read error:", err)
			}
			return
		}

		switch opcode {
		case 1: // текстовый фрейм
			var generic struct {
				Op    string `json:"op"`
				Topic string `json:"topic"`
			}
			if err := json.Unmarshal(frame, &generic); err != nil {
				fmt.Println("[bybit] unmarshal error:", err)
				continue
			}

			switch {
			case generic.Op == "pong":
				fmt.Println("[bybit] pong received")
			case generic.Topic == "trade.BTCUSDT":
				b.handleTradeMessage(frame, out)
			default:
				// игнорируем
			}

		case 8: // Close frame
			fmt.Println("[bybit] close frame received, sending close frame back and closing connection")
			// отправить close frame в ответ
			if err := writeCloseFrame(b.getConn()); err != nil {
				fmt.Println("[bybit] error sending close frame:", err)
			}
			return // выходим из readLoop, чтобы перезапустить соединение

		case 9: // ping
			fmt.Println("[bybit] ping received, sending pong")
			if err := writePongFrame(conn); err != nil {
				fmt.Println("[bybit] pong write error:", err)
				return
			}

		case 10: // pong
			fmt.Println("[bybit] pong received")

		default:
			fmt.Printf("[bybit] ignoring frame opcode %d\n", opcode)
		}
	}
}

func (b *BybitAdapter) handleTradeMessage(frame []byte, out chan<- domain.MarketData) {
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
		fmt.Println("[bybit] trade message unmarshal error:", err)
		return
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
			return
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
		case <-ticker.C:
			conn := b.getConn()
			if conn == nil {
				fmt.Println("[bybit] pingLoop: connection is nil, stopping ping loop")
				return
			}
			if err := writePingFrameBybit(conn); err != nil {
				fmt.Println("[bybit] ping write error:", err)
				return // прерываем для реконнекта
			}
		}
	}
}

func writePingFrameBybit(conn *tls.Conn) error {
	pingMsg := map[string]string{"op": "ping"}
	msgJSON, err := json.Marshal(pingMsg)
	if err != nil {
		return err
	}
	return writeFrame(conn, msgJSON)
}

// readFrame читает один WebSocket фрейм (текстовый или control frame)
func readFrame(r io.Reader) (opcode byte, payload []byte, err error) {
	var header [2]byte
	if _, err = io.ReadFull(r, header[:]); err != nil {
		return
	}

	fin := header[0]&0x80 != 0
	opcode = header[0] & 0x0F
	mask := header[1]&0x80 != 0
	payloadLen := int(header[1] & 0x7F)

	if payloadLen == 126 {
		var extendedLen uint16
		err = binary.Read(r, binary.BigEndian, &extendedLen)
		if err != nil {
			return
		}
		payloadLen = int(extendedLen)
	} else if payloadLen == 127 {
		var extendedLen uint64
		err = binary.Read(r, binary.BigEndian, &extendedLen)
		if err != nil {
			return
		}
		if extendedLen > (1 << 31) {
			err = errors.New("payload too large")
			return
		}
		payloadLen = int(extendedLen)
	}

	var maskingKey [4]byte
	if mask {
		if _, err = io.ReadFull(r, maskingKey[:]); err != nil {
			return
		}
	}

	payload = make([]byte, payloadLen)
	if _, err = io.ReadFull(r, payload); err != nil {
		return
	}

	if mask {
		for i := 0; i < payloadLen; i++ {
			payload[i] ^= maskingKey[i%4]
		}
	}

	if !fin {
		err = errors.New("fragmented frames not supported")
		return
	}

	return opcode, payload, nil
}

func writeFrame(conn io.Writer, data []byte) error {
	// Формируем заголовок фрейма WebSocket для текстового фрейма (opcode = 0x1)
	// FIN = 1 (последний фрейм)
	// MASK = 0 (сервер не маскирует данные)

	payloadLen := len(data)
	var header []byte

	if payloadLen <= 125 {
		header = []byte{0x81, byte(payloadLen)} // 0x81 = FIN + текстовый фрейм
	} else if payloadLen <= 65535 {
		header = make([]byte, 4)
		header[0] = 0x81
		header[1] = 126
		binary.BigEndian.PutUint16(header[2:], uint16(payloadLen))
	} else {
		header = make([]byte, 10)
		header[0] = 0x81
		header[1] = 127
		binary.BigEndian.PutUint64(header[2:], uint64(payloadLen))
	}

	// Записываем заголовок
	if _, err := conn.Write(header); err != nil {
		return err
	}

	// Записываем данные
	_, err := conn.Write(data)
	return err
}

func writePongFrame(conn io.Writer) error {
	// Pong — это control frame opcode = 0xA, без payload (или с payload)
	header := []byte{0x8A, 0} // FIN=1, opcode=0xA, длина payload=0
	_, err := conn.Write(header)
	return err
}

func writeCloseFrame(conn io.Writer) error {
	// Close frame opcode = 0x8, без payload
	header := []byte{0x88, 0} // FIN=1, opcode=0x8, длина payload=0
	_, err := conn.Write(header)
	return err
}
