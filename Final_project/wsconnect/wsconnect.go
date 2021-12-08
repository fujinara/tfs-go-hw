package wsconnect

import (
	"context"
	"encoding/json"
	"final_project/domain"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Event      string   `json:"event"`
	Feed       string   `json:"feed"`
	ProductIDs []string `json:"product_ids"`
}

func Subscribe(ctx context.Context, wg *sync.WaitGroup, ticker string) <-chan domain.Price {
	prices := make(chan domain.Price)
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(prices)

		u := "wss://demo-futures.kraken.com/ws/v1"
		var conn *websocket.Conn
		var err error
		success := false
		for i := 0; i < 5; i++ {
			log.Printf("connecting to %s", u)
			conn, _, err = websocket.DefaultDialer.Dial(u, nil)
			if err != nil {
				log.Println("dial error: ", err)
				time.Sleep(3 * time.Second)
				log.Println("attempting to reconnect...")
			} else {
				defer conn.Close()
				log.Println("connection has been established")
				success = true
				break
			}
		}
		if !success {
			log.Println("connection could not be established")
			return
		}

		var msg = Message{"subscribe", "ticker", []string{ticker}}
		err = conn.WriteJSON(msg)
		if err != nil {
			log.Println("write: ", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, data, err := conn.ReadMessage()
				if err != nil {
					log.Println("read: ", err)
				}
				jsonData := make(map[string]interface{})
				_ = json.Unmarshal(data, &jsonData)
				p := domain.Price{}
				p.Ticker, _ = jsonData["product_id"].(string)
				p.Bid, _ = jsonData["bid"].(float64)
				p.Ask, _ = jsonData["ask"].(float64)
				if (domain.Price{}) != p {
					fmt.Println(p)
					prices <- p
				}
			}
		}
	}()

	return prices
}
