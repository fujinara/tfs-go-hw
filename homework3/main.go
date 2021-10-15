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

	go func() {
		<-c
		logger.Info("interruption signal is recieved")
		cancel()
	}()
	
	prices := pg.Prices(ctx, &wg)
	candle1mToSave := converter.PricesToCandle(prices, &wg)
	candle1m := converter.SaveCandleCSV(candle1mToSave, domain.CandlePeriod1m, &wg)
	candle2mToSave := converter.CanldeToCandle(candle1m, &wg)
	candle2m := converter.SaveCandleCSV(candle2mToSave, domain.CandlePeriod2m, &wg)
	candle10mToSave := converter.CanldeToCandle(candle2m, &wg)
	candles10m := converter.SaveCandleCSV(candle10mToSave, domain.CandlePeriod10m, &wg)
	for _ = range candles10m {}

	wg.Wait()
}
