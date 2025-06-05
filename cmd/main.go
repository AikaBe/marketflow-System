package main

import (
	"context"
	"fmt"
	"marketflow/internal/adapters/websocket/impl"
	"marketflow/internal/usecase"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	binance := impl.NewBinanceAdapter()
	bybit := impl.NewBybitAdapter()

	fmt.Println("Binance and Bybit adapters created")

	service := usecase.NewMarketDataService(
		binance,
		bybit,
	)

	dataChan := service.Start(ctx)

	fmt.Println("Market data service started")

	go func() {
		for data := range dataChan {
			fmt.Printf("[%s] %s: %.2f (%.2f) at %d\n",
				data.Exchange, data.Symbol, data.Price, data.Volume, data.Time)
		}
	}()

	<-ctx.Done()
	fmt.Println("Stopping market data service...")
	service.Stop()
	fmt.Println("Graceful shutdown")
}
