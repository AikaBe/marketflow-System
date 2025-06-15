package tar_files

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
			log.Printf("[%s] Error connecting: %v", name, err)
			time.Sleep(2 * time.Second)
			continue
		}
		log.Printf("[%s] Connected to %s", name, address)
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			var t Ticker
			if err := json.Unmarshal([]byte(scanner.Text()), &t); err != nil {
				continue
			}
			out <- domain.PriceUpdate{
				Symbol:    t.Symbol,
				Price:     t.Price,
				Timestamp: t.Timestamp,
				Exchange:  name,
			}
		}
		conn.Close()
		time.Sleep(2 * time.Second)
	}
}

func StartReaders(out chan<- domain.PriceUpdate) {
	go connectAndRead("Exchange1", "exchange1:40101", out)
	go connectAndRead("Exchange2", "exchange2:40102", out)
	go connectAndRead("Exchange3", "exchange3:40103", out)
}
