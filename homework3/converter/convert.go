package converter

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"hw-async/domain"
	"math"
	"os"
	"sync"
)

type CurrState map[string]*domain.Candle

func PricesToCandle(prices <-chan domain.Price, ctx context.Context, wg *sync.WaitGroup) <-chan domain.Candle {
	candles1m := make(chan domain.Candle)
	wg.Add(1)
	go func() {
		defer wg.Done()
		curr := make(CurrState)
		for {
			select {
			case <-ctx.Done():
				return
			case <-prices:
				for entry := range prices {
					if _, ok := curr[entry.Ticker]; ok {
						curr[entry.Ticker].High = math.Max(curr[entry.Ticker].High, entry.Value)
						curr[entry.Ticker].Low = math.Min(curr[entry.Ticker].Low, entry.Value)
						curr[entry.Ticker].Close = entry.Value
						timestamp, err := domain.PeriodTS(domain.CandlePeriod1m, entry.TS)
						if err != nil {
							panic(err)
						}
						if curr[entry.Ticker].TS != timestamp {
							candles1m <- *curr[entry.Ticker]
							curr[entry.Ticker] = &domain.Candle{
								Ticker: entry.Ticker,
								Period: domain.CandlePeriod1m,
								Open:   entry.Value,
								High:   entry.Value,
								Low:    entry.Value,
								Close:  entry.Value,
								TS:     timestamp,
							}
						}
					} else {
						timestamp, err := domain.PeriodTS(domain.CandlePeriod1m, entry.TS)
						if err != nil {
							panic(err)
						}
						curr[entry.Ticker] = &domain.Candle{
							Ticker: entry.Ticker,
							Period: domain.CandlePeriod1m,
							Open:   entry.Value,
							High:   entry.Value,
							Low:    entry.Value,
							Close:  entry.Value,
							TS:     timestamp,
						}
					}
				}
			}
		}
	}()
	return candles1m
}

func CanldeToCandle(candlesIn <-chan domain.Candle, ctx context.Context, wg *sync.WaitGroup) <-chan domain.Candle {
	candlesOut := make(chan domain.Candle)
	wg.Add(1)
	go func() {
		defer wg.Done()
		curr := make(CurrState)
		for {
			select {
			case <-ctx.Done():
				return
			case <-candlesIn:
				for cand := range candlesIn {
					if _, ok := curr[cand.Ticker]; ok {
						curr[cand.Ticker].High = math.Max(curr[cand.Ticker].High, cand.High)
						curr[cand.Ticker].Low = math.Min(curr[cand.Ticker].Low, cand.Low)
						curr[cand.Ticker].Close = cand.Close
						var newPeriod domain.CandlePeriod
						switch cand.Period {
						case domain.CandlePeriod1m:
							newPeriod = domain.CandlePeriod2m
						case domain.CandlePeriod2m:
							newPeriod = domain.CandlePeriod10m
						default:
							panic(errors.New("unknown period"))
						}
						timestamp, err := domain.PeriodTS(newPeriod, cand.TS)
						if err != nil {
							panic(err)
						}
						if curr[cand.Ticker].TS != timestamp {
							candlesOut <- *curr[cand.Ticker]
							curr[cand.Ticker] = &domain.Candle{
								Ticker: cand.Ticker,
								Period: newPeriod,
								Open:   cand.Open,
								High:   cand.High,
								Low:    cand.Low,
								Close:  cand.Close,
								TS:     timestamp,
							}
						}
					} else {
						var newPeriod domain.CandlePeriod
						switch cand.Period {
						case domain.CandlePeriod1m:
							newPeriod = domain.CandlePeriod2m
						case domain.CandlePeriod2m:
							newPeriod = domain.CandlePeriod10m
						default:
							panic(errors.New("unknown period"))
						}
						timestamp, err := domain.PeriodTS(newPeriod, cand.TS)
						if err != nil {
							panic(err)
						}
						curr[cand.Ticker] = &domain.Candle{
							Ticker: cand.Ticker,
							Period: newPeriod,
							Open:   cand.Open,
							High:   cand.High,
							Low:    cand.Low,
							Close:  cand.Close,
							TS:     timestamp,
						}
					}
				}
			}
		}
	}()
	return candlesOut
}

func CandleToStr(cand domain.Candle) []string {
	return []string{cand.Ticker, cand.TS.String(), fmt.Sprintf("%f", cand.Open), fmt.Sprintf("%f", cand.High), fmt.Sprintf("%f", cand.Low), fmt.Sprintf("%f", cand.Close)}
}

func SaveCandleCSV(candlesIn <-chan domain.Candle, period domain.CandlePeriod, ctx context.Context, wg *sync.WaitGroup) <-chan domain.Candle {
	candlesOut := make(chan domain.Candle)
	filename := fmt.Sprintf("candles_%s.csv", period)
	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(f)
	defer w.Flush()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-candlesIn:
				for cand := range candlesIn {
					er := w.Write(CandleToStr(cand))
					if er != nil {
						panic(er)
					}
					candlesOut <- cand
				}
			}
		}
	}()
	return candlesOut
}
