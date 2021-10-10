package main

import (
	"context"
	"hw-async/generator"
	"hw-async/domain"
	"sync"
	"os"
	"os/signal"
	"time"
	"hw-async/converter"

	log "github.com/sirupsen/logrus"
)

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	logger := log.New()
	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-c:
			logger.Info("interruption signal is recieved")
			cancel()
		case <-ctx.Done():
		}
	}()
	
	prices := pg.Prices(ctx, &wg)
	candle1mToSave := converter.PricesToCandle(prices, ctx, &wg)
	candle1m := converter.SaveCandleCSV(candle1mToSave, domain.CandlePeriod1m, ctx, &wg)
	candle2mToSave := converter.CanldeToCandle(candle1m, ctx, &wg)
	candle2m := converter.SaveCandleCSV(candle2mToSave, domain.CandlePeriod2m, ctx, &wg)
	candle10mToSave := converter.CanldeToCandle(candle2m, ctx, &wg)
	_ = converter.SaveCandleCSV(candle10mToSave, domain.CandlePeriod10m, ctx, &wg)

	wg.Wait()
}
