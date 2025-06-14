package tar_files

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// Ticker — структура одного сообщения
type Ticker struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

// connectAndRead подключается к адресу и выводит полученные JSON-объекты
func connectAndRead(name, address string) {
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
			err := json.Unmarshal([]byte(line), &t)
			if err != nil {
				log.Printf("[%s] Failed to parse JSON: %s", name, line)
				continue
			}
			fmt.Printf("[%s] %s | %.4f | %d\n", name, t.Symbol, t.Price, t.Timestamp)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("[%s] Connection error: %v", name, err)
		}
		conn.Close()
		time.Sleep(2 * time.Second) // попробуем переподключиться
	}
}

// StartReaders запускает параллельно чтение из всех источников
func StartReaders() {
	go connectAndRead("Exchange1", "exchange1:40101")
	go connectAndRead("Exchange2", "exchange2:40102")
	go connectAndRead("Exchange3", "exchange3:40103")
}
