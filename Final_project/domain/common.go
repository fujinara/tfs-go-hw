package domain

import "time"

const (
	Layout        = "2006-01-02T15:04:05Z0700"
	TgUserID      = int64(95531009)
	ExchangeURL   = "https://demo-futures.kraken.com/derivatives"
	SendOrderEndp = "/api/v3/sendorder"
	OpenPosEndp   = "/api/v3/openpositions"
)

type OrderRequest struct {
	Side   string
	Ticker string
	Size   int64
	Price  float64
}

type StrategyRequest struct {
	Ticker     string
	Size       int64
	Percentage float64
}

type Price struct {
	Ticker string
	Bid    float64
	Ask    float64
}

type SendOrderResp struct {
	Result     string         `json:"result,omitempty"`
	SendStatus SendStatusType `json:"sendStatus,omitempty"`
	ServerTime interface{}    `json:"ServerTime,omitempty"`
	Error      string         `json:"error,omitempty"`
}

type SendStatusType struct {
	Order_id     string            `json:"order_id,omitempty"`
	Status       string            `json:"status,omitempty"`
	ReceivedTime interface{}       `json:"receivedTime,omitempty"`
	OrderEvents  []OrderEventsType `json:"orderEvents,omitempty"`
}

type OrderEventsType struct {
	ExecutionId          string                  `json:"executionId,omitempty"`
	Uid                  string                  `json:"uid,omitempty"`
	Price                float64                 `json:"price,omitempty"`
	Amount               int64                   `json:"amount,omitempty"`
	Order                OrderPriorExecutionType `json:"order,omitempty"`
	OrderPriorEdit       interface{}             `json:"orderPriorEdit,omitempty"`
	OrderPriorExecution  OrderPriorExecutionType `json:"OrderPriorExecution,omitempty"`
	TakerReducedQuantity interface{}             `json:"takerReducedQuantity,omitempty"`
	Reason               string                  `json:"reason,omitempty"`
	Type                 string                  `json:"type,omitempty"`
}

type OrderPriorExecutionType struct {
	OrderId             string      `json:"orderId,omitempty"`
	CliOrdId            interface{} `json:"cliOrdId,omitempty"`
	Type                string      `json:"type,omitempty"`
	Symbol              string      `json:"symbol,omitempty"`
	Side                string      `json:"side,omitempty"`
	Quantity            float64     `json:"quantity,omitempty"`
	Filled              float64     `json:"filled,omitempty"`
	LimitPrice          float64     `json:"limitPrice,omitempty"`
	ReduceOnly          bool        `json:"reduceOnly,omitempty"`
	Timestamp           interface{} `json:"timestamp,omitempty"`
	LastUpdateTimestamp interface{} `json:"lastUpdateTimestamp,omitempty"`
}

type OpenPositionResp struct {
	Result        string         `json:"result,omitempty"`
	OpenPositions []OpenPosition `json:"openPositions,omitempty"`
	ServerTime    interface{}    `json:"serverTime,omitempty"`
	Error         string         `json:"error,omitempty"`
}

type OpenPosition struct {
	Side              string      `json:"side,omitempty"`
	Symbol            string      `json:"symbol,omitempty"`
	Price             float64     `json:"price,omitempty"`
	FillTime          interface{} `json:"fillTime,omitempty"`
	Size              int64       `json:"size,omitempty"`
	UnrealizedFunding float64     `json:"unrealizedFunding,omitempty"`
}

// take profit/stop loss response
type TPSLResponse struct {
	OrderId    string
	Status     string
	Instrument string
	Type       string
	Size       int64
	OpenPrice  float64
	ClosePrice float64
	Profit     float64
	TS         time.Time
}
