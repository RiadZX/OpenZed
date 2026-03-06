package utils

import (
	"encoding/json"
	"github.com/gookit/slog"
	"net/http"
	"time"
)

var (
	//unix timestamp
	lastUpdated int64 = 0
	solPrice          = 220.0
)

type RespPrice struct {
	Solana struct {
		Usd float64 `json:"usd"`
	} `json:"solana"`
}

func fetchSolPrice() int {
	//fetch the sol price and set the last updated time
	lastUpdated = time.Now().Unix()
	//set the sol price, from an api
	url := "https://api.coingecko.com/api/v3/simple/price?ids=solana&vs_currencies=usd"
	//fetch the price from the api
	resp, err := http.Get(url)
	if err != nil {
		return 220
	}
	defer resp.Body.Close()
	//parse the response
	var price RespPrice
	err = json.NewDecoder(resp.Body).Decode(&price)
	if err != nil {
		return 220
	}
	solPrice = price.Solana.Usd
	slog.Debug("Got the price: ", solPrice)
	return int(solPrice)
}

func GetSolPrice() int {
	if lastUpdated == 0 {
		//get the price since this is the first time
		return fetchSolPrice()
	}
	//if already loaded, check if it is 30min old, compare the unix timestamp
	if time.Now().Unix()-lastUpdated > 1800 {
		//get the price since it is 30min old
		return fetchSolPrice()
	}
	return int(solPrice)
}
