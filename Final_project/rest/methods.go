package rest

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type Kraken struct {
	publicKey  string
	privateKey string
}

// возвращать ошибки
func New(apikey string, secret string) *Kraken {
	if apikey == "" || secret == "" {
		log.Fatal("warning: either api key or secret is empty")
	}
	return &Kraken{
		publicKey:  apikey,
		privateKey: secret,
	}
}

func (api *Kraken) GetAuthent(postData, endpointPath string) string {
	sha := sha256.New()
	src := postData + endpointPath
	sha.Write([]byte(src))

	apiDecode, _ := base64.StdEncoding.DecodeString(api.privateKey)

	h := hmac.New(sha512.New, apiDecode)
	h.Write(sha.Sum(nil))

	result := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return result
}

func (api *Kraken) SendOrder(side string, symbol string, size int64, price float64, exchurl string, endp string) (*http.Response, error) {
	params := url.Values{}
	params.Add("orderType", "ioc")
	params.Add("side", side)
	params.Add("symbol", symbol)
	params.Add("size", strconv.FormatInt(size, 10))
	params.Add("limitPrice", strconv.FormatFloat(price, 'f', 2, 64))

	SendOrderUrl := exchurl + endp
	u, _ := url.ParseRequestURI(SendOrderUrl)
	postData := params.Encode()
	u.RawQuery = postData

	urlStr := fmt.Sprintf("%v", u)

	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return nil, err
	}
	signature := api.GetAuthent(postData, endp)
	req.Header.Add("Authent", signature)
	req.Header.Add("APIKey", api.publicKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (api *Kraken) GetLastOpenPosition(instrument string, exchurl string, endp string) (*http.Response, error) {
	u := exchurl + endp

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	signature := api.GetAuthent("", endp)
	req.Header.Add("Authent", signature)
	req.Header.Add("APIKey", api.publicKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
