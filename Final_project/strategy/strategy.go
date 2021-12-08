package strategy

import (
	"context"
	"encoding/json"
	"errors"
	"final_project/domain"
	"final_project/rest"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/jackc/pgx/v4"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TakeProfitStopLoss(prices <-chan domain.Price, k *rest.Kraken, t *tgbotapi.BotAPI, conn *pgx.Conn, instrument string, size int64, percentage float64) {
	resp, err := k.GetLastOpenPosition(instrument, domain.ExchangeURL, domain.OpenPosEndp)
	if err != nil {
		log.Println(err)
		return
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var jsonData domain.OpenPositionResp
	_ = json.Unmarshal(respData, &jsonData)

	var lastOP domain.OpenPosition
	if jsonData.Result == "error" {
		log.Println(errors.New(jsonData.Error))
		return
	} else if len(jsonData.OpenPositions) == 0 {
		log.Println(errors.New("no open positions"))
		return
	} else {
		for _, elem := range jsonData.OpenPositions {
			if elem.Symbol == instrument {
				lastOP = elem
				break
			}
		}
		if lastOP == (domain.OpenPosition{}) {
			log.Println(errors.New("no open positions of a given ticker"))
		}
	}

	profitPrice := lastOP.Price * (1 + percentage/100)
	lossPrice := lastOP.Price * (1 - percentage/100)
	if size > lastOP.Size {
		log.Println(errors.New("given size is greater than the open position size"))
		return
	}

	if lastOP.Side == "long" {
		for entry := range prices {
			if entry.Bid >= profitPrice || entry.Bid <= lossPrice {
				resp, err := k.SendOrder("sell", lastOP.Symbol, size, entry.Bid, domain.ExchangeURL, domain.SendOrderEndp)
				if err != nil {
					log.Println(err)
					return
				}
				respData, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					return
				}
				resp.Body.Close()
				var jsonData domain.SendOrderResp
				_ = json.Unmarshal(respData, &jsonData)
				if jsonData.Result == "error" {
					log.Println(errors.New(jsonData.Error))
					return
				}

				if jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Type == "EXECUTION" {
					ts, err := time.Parse(domain.Layout, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].OrderPriorExecution.Timestamp.(string))
					if err != nil {
						log.Println(err)
					}
					execInfo := domain.TPSLResponse{
						OrderId:    jsonData.SendStatus.Order_id,
						Status:     jsonData.SendStatus.Status,
						Instrument: jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].OrderPriorExecution.Symbol,
						Type:       "long",
						Size:       jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Amount,
						OpenPrice:  lastOP.Price,
						ClosePrice: jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Price,
						Profit:     jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Price - lastOP.Price,
						TS:         ts,
					}
					text := fmt.Sprintf("Position was closed\nOrderID : %v\nStatus : %v\nInstrument : %v\nType : %v\nSize : %v\nOpen Price : %v\nClose Price : %v\nProfit : %v\nTime : %v\n", execInfo.OrderId, execInfo.Status, execInfo.Instrument, execInfo.Type, execInfo.Size, execInfo.OpenPrice, execInfo.ClosePrice, execInfo.Profit, execInfo.TS)
					msg := tgbotapi.NewMessage(domain.TgUserID, text)
					t.Send(msg)
					conn.Exec(context.Background(), "INSERT INTO ordersinfo(orderID, instrument, price, size, order_type, ts) VALUES($1, $2, $3, $4, $5, $6)", execInfo.OrderId, execInfo.Instrument, execInfo.ClosePrice, execInfo.Size, "sell", execInfo.TS)
				} else if jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Type == "REJECT" {
					ts, err := time.Parse(domain.Layout, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Order.Timestamp.(string))
					if err != nil {
						log.Println(err)
					}
					execInfo := domain.TPSLResponse{
						OrderId:    jsonData.SendStatus.Order_id,
						Status:     jsonData.SendStatus.Status,
						Instrument: jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Order.Symbol,
						Type:       "long",
						TS:         ts,
					}
					fmt.Println("The order has been rejected")
					fmt.Println("OrderID : ", execInfo.OrderId)
					fmt.Println("Reason : ", execInfo.Status)
					fmt.Println("Instrument : ", execInfo.Instrument)
					fmt.Println("Type : ", execInfo.Type)
					fmt.Println("Time : ", execInfo.TS)
					fmt.Println("======================================================")
				} else {
					log.Println(errors.New("unknown order status"))
					return
				}
				return
			}
		}
	} else {
		for entry := range prices {
			if entry.Ask >= profitPrice || entry.Ask <= lossPrice {
				resp, err := k.SendOrder("buy", lastOP.Symbol, size, entry.Ask, domain.ExchangeURL, domain.SendOrderEndp)
				if err != nil {
					log.Println(err)
					return
				}
				respData, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					return
				}
				resp.Body.Close()
				var jsonData domain.SendOrderResp
				_ = json.Unmarshal(respData, &jsonData)
				if jsonData.Result == "error" {
					log.Println(errors.New(jsonData.Error))
					return
				}

				if jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Type == "EXECUTION" {
					ts, err := time.Parse(domain.Layout, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].OrderPriorExecution.Timestamp.(string))
					if err != nil {
						log.Println(err)
					}
					execInfo := domain.TPSLResponse{
						OrderId:    jsonData.SendStatus.Order_id,
						Status:     jsonData.SendStatus.Status,
						Instrument: jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].OrderPriorExecution.Symbol,
						Type:       "short",
						Size:       jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Amount,
						OpenPrice:  lastOP.Price,
						ClosePrice: jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Price,
						Profit:     lastOP.Price - jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Price,
						TS:         ts,
					}
					text := fmt.Sprintf("Position was closed\nOrderID : %v\nStatus : %v\nInstrument : %v\nType : %v\nSize : %v\nOpen Price : %v\nClose Price : %v\nProfit : %v\nTime : %v\n", execInfo.OrderId, execInfo.Status, execInfo.Instrument, execInfo.Type, execInfo.Size, execInfo.OpenPrice, execInfo.ClosePrice, execInfo.Profit, execInfo.TS)
					msg := tgbotapi.NewMessage(domain.TgUserID, text)
					t.Send(msg)
					conn.Exec(context.Background(), "INSERT INTO ordersinfo(orderID, instrument, price, size, order_type, ts) VALUES($1, $2, $3, $4, $5, $6)", execInfo.OrderId, execInfo.Instrument, execInfo.ClosePrice, execInfo.Size, "buy", execInfo.TS)
				} else if jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Type == "REJECT" {
					ts, err := time.Parse(domain.Layout, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Order.Timestamp.(string))
					if err != nil {
						log.Println(err)
					}
					execInfo := domain.TPSLResponse{
						OrderId:    jsonData.SendStatus.Order_id,
						Status:     jsonData.SendStatus.Status,
						Instrument: jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Order.Symbol,
						Type:       "short",
						TS:         ts,
					}
					fmt.Println("The order has been rejected")
					fmt.Println("OrderID : ", execInfo.OrderId)
					fmt.Println("Reason : ", execInfo.Status)
					fmt.Println("Instrument : ", execInfo.Instrument)
					fmt.Println("Type : ", execInfo.Type)
					fmt.Println("Time : ", execInfo.TS)
					fmt.Println("======================================================")
				} else {
					log.Println(errors.New("unknown order status"))
					return
				}
				return
			}
		}
	}
}
