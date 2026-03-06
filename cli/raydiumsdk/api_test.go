package raydiumsdk

import (
	"github.com/gagliardetto/solana-go/rpc"
	"testing"
)

func TestGetPricePerToken(t *testing.T) {
	client := rpc.New("https://api.mainnet-beta.solana.com")
	address := "2QJN63Z9v4aWsEBgCZgF8sSZk5MbwFWAz6c93J9udxnQ"
	delay := int64(1000)
	retries := int64(3)
	pricePerToken, baseIsSol, mint, decimalsMint, supply, err := GetPricePerToken(client, address, delay, retries)
	if err != nil {
		t.Error(err)
	}
	t.Log(pricePerToken, baseIsSol, mint, decimalsMint, supply)
}
