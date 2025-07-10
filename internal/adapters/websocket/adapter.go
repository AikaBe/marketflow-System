package websocket

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"net"
	"time"

	"marketflow/internal/domain"
)

type Ticker struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

// Fan-In pattern
func connectAndRead(name, address string, out chan<- domain.PriceUpdate) {
	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			slog.Error("Connection failed",
				"exchange", name,
				"address", address,
				"err", err,
			)
			time.Sleep(2 * time.Second)
			continue
		}
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			var t Ticker
			line := scanner.Text()
			if err := json.Unmarshal([]byte(line), &t); err != nil {
				slog.Error("Failed to parse ticker JSON",
					"exchange", name,
					"raw", line,
					"err", err,
				)
				continue
			}

			out <- domain.PriceUpdate{
				Symbol:    t.Symbol,
				Price:     t.Price,
				Timestamp: t.Timestamp,
				Exchange:  name,
			}
		}

		if err := scanner.Err(); err != nil {
			slog.Warn("Scanner error occurred",
				"exchange", name,
				"err", err,
			)
		}
		conn.Close()
		time.Sleep(2 * time.Second)
	}
}

func StartReaders(out chan<- domain.PriceUpdate) {
	slog.Info("[LIVE MODE] Starting WebSocket Readers...")
	go connectAndRead("Exchange1", "exchange1:40101", out)
	go connectAndRead("Exchange2", "exchange2:40102", out)
	go connectAndRead("Exchange3", "exchange3:40103", out)
}
