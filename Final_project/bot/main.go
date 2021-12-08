package main

import (
	"context"
	"encoding/json"
	"errors"
	"final_project/domain"
	"final_project/rest"
	"final_project/strategy"
	"final_project/wsconnect"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	krk   rest.Kraken
	tgbot tgbotapi.BotAPI
	conn  pgx.Conn
}

func NewClient(krk *rest.Kraken, tgbot *tgbotapi.BotAPI, conn *pgx.Conn) *Client {
	return &Client{
		krk:   *krk,
		tgbot: *tgbot,
		conn:  *conn,
	}
}

func main() {
	root := chi.NewRouter()
	root.Use(middleware.Logger)
	k := rest.New(os.Getenv("KRAKEN_API_KEY"), os.Getenv("KRAKEN_SECRET"))
	t, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := pgx.Connect(context.Background(), "postgres://postgres:@localhost:5432/postgres")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	if err = conn.Ping(context.Background()); err != nil {
		log.Fatalf("can't ping db: %s", err)
	}

	h := NewClient(k, t, conn)
	root.Post("/sendorder", h.Order)
	root.Post("/strategy", h.Algorithm)

	log.Fatal(http.ListenAndServe(":5000", root))
}

func (h *Client) Order(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var m domain.OrderRequest
	err = json.Unmarshal(d, &m)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.krk.SendOrder(m.Side, m.Ticker, m.Size, m.Price, domain.ExchangeURL, domain.SendOrderEndp)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}
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
		id, status, instrument, side, size, price := jsonData.SendStatus.Order_id, jsonData.SendStatus.Status, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].OrderPriorExecution.Symbol, m.Side, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Amount, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Price
		text := fmt.Sprintf("Position was opened\nOrderID : %v\nStatus : %v\nInstrument : %v\nSide : %v\nSize : %v\nPrice : %v\nTime : %v\n", id, status, instrument, side, size, price, ts)
		msg := tgbotapi.NewMessage(domain.TgUserID, text)
		h.tgbot.Send(msg)
		h.conn.Exec(context.Background(), "INSERT INTO ordersinfo(orderID, instrument, price, size, order_type, ts) VALUES($1, $2, $3, $4, $5, $6)", id, instrument, price, size, side, ts)
		log.Println("The order was placed")
		return
	} else {
		log.Println("The order was not executed. Reason : ", jsonData.SendStatus.Status)
		return
	}
}

func (h *Client) Algorithm(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var m domain.StrategyRequest
	err = json.Unmarshal(d, &m)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	prices := wsconnect.Subscribe(ctx, &wg, m.Ticker)
	strategy.TakeProfitStopLoss(prices, &h.krk, &h.tgbot, &h.conn, m.Ticker, m.Size, m.Percentage)
	cancel()
	for range prices {
	}
	wg.Wait()
}
