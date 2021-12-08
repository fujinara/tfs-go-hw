package rest

import (
	"encoding/json"
	"errors"
	"final_project/domain"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success","sendStatus":{"order_id":"f0cfe4af-8e24-4b8e-81ab-95cc778e2ed3","status":"placed","receivedTime":"2021-12-03T17:10:27.546Z","orderEvents":[{"executionId":"51452405-34fa-4b02-ae9d-5c91dae8d4c4","price":54969.50,"amount":1,"orderPriorEdit":null,"orderPriorExecution":{"orderId":"f0cfe4af-8e24-4b8e-81ab-95cc778e2ed3","cliOrdId":null,"type":"ioc","symbol":"pi_xbtusd","side":"buy","quantity":1,"filled":0,"limitPrice":67830.00,"reduceOnly":false,"timestamp":"2021-12-03T17:10:27.546Z","lastUpdateTimestamp":"2021-12-03T17:10:27.546Z"},"takerReducedQuantity":null,"type":"EXECUTION"}]},"serverTime":"2021-12-03T17:10:27.548Z"}`))
	}))

	kr := Kraken{
		publicKey:  "some",
		privateKey: "keys",
	}

	resp, err := kr.SendOrder("buy", "pi_xbtusd", 1, 67830.00, server.URL, "")

	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	var jsonData domain.SendOrderResp
	_ = json.Unmarshal(respData, &jsonData)
	if jsonData.Result == "error" {
		t.Error(errors.New(jsonData.Error))
	}

	id, status, instrument, side, size, price := jsonData.SendStatus.Order_id, jsonData.SendStatus.Status, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].OrderPriorExecution.Symbol, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].OrderPriorExecution.Side, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Amount, jsonData.SendStatus.OrderEvents[len(jsonData.SendStatus.OrderEvents)-1].Price

	if id != "f0cfe4af-8e24-4b8e-81ab-95cc778e2ed3" || status != "placed" || instrument != "pi_xbtusd" || side != "buy" || size != 1 || price != 54969.50 {
		t.Error(errors.New("error while parsing response from server"))
	}
}

func TestGetLastOpenPosition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success", "openPositions":[{"side":"short", "symbol":"pi_xbtusd", "price":9392.75, "fillTime":"2020-07-22T14:39:12.376Z", "size":10000, "unrealizedFunding":1.045432180096817E-5}], "serverTime":"2020-07-22T14:39:12.376Z"}`))
	}))

	kr := Kraken{
		publicKey:  "some",
		privateKey: "keys",
	}

	resp, err := kr.GetLastOpenPosition("pi_xbtusd", server.URL, "")
	if err != nil {
		t.Error(err)
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	var jsonData domain.OpenPositionResp
	_ = json.Unmarshal(respData, &jsonData)
	var lastOP domain.OpenPosition
	if jsonData.Result == "error" {
		t.Error(errors.New(jsonData.Error))
	} else if len(jsonData.OpenPositions) == 0 {
		t.Error(errors.New("no open positions"))
	} else {
		for _, elem := range jsonData.OpenPositions {
			if elem.Symbol == "pi_xbtusd" {
				lastOP = elem
				break
			}
		}
		if lastOP == (domain.OpenPosition{}) {
			t.Error(errors.New("no open positions of a given ticker"))
		}
	}

	if lastOP.Side != "short" || lastOP.Price != 9392.75 || lastOP.Size != 10000 || lastOP.Symbol != "pi_xbtusd" {
		t.Error(errors.New("error while parsing response from server"))
	}
}
