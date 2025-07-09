package websocket

import (
	"bufio"
	"encoding/json"
	"log"
	"marketflow/internal/domain"
	"net"
	"time"
)

type Ticker struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

// Fan-In
func connectAndRead(name, address string, out chan<- domain.PriceUpdate) {
	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Printf("[%s] Error connecting to %s: %v", name, address, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("[%s] Connected to %s", name, address)
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			var t Ticker
			line := scanner.Text()
			if err := json.Unmarshal([]byte(line), &t); err != nil {
				log.Printf("[%s] Failed to parse JSON: %s | err: %v", name, line, err)
				continue
			}

			out <- domain.PriceUpdate{
				Symbol:    t.Symbol,
				Price:     t.Price,
				Timestamp: t.Timestamp,
				Exchange:  name,
			}
		}

		log.Printf("[%s] Disconnected from %s, retrying...", name, address)
		conn.Close()
		time.Sleep(2 * time.Second)
	}
}

func StartReaders(out chan<- domain.PriceUpdate) {
	log.Println("[LIVE MODE] Starting WebSocket Readers...")
	go connectAndRead("Exchange1", "exchange1:40101", out)
	go connectAndRead("Exchange2", "exchange2:40102", out)
	go connectAndRead("Exchange3", "exchange3:40103", out)
}
