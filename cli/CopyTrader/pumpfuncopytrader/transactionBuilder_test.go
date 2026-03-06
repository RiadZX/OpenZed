package pumpfuncopytrader

import (
	"Zed/utils"
	"testing"
)

func TestConstructCreateATAInstruction(t *testing.T) {
	var userPublicKey = "userPublicKey"
	var tokenMint = "tokenMint"

	_, err := utils.ConstructCreateATAInstruction(userPublicKey, tokenMint)
	if err == nil {
		t.Errorf("Test failed, expected: %v, got: %v", "decode: invalid base58 digit ('l')", err)
	}

	userPublicKey = "ZDK6SbnjEtgTWG1CqHhNueEFvziCcMQYQhW2UwJ3EzU"
	tokenMint = "EeJ1TH7MACdR8dmHBaGGrkZrsHqWG86J74F9m5fZpump"
	_, err = utils.ConstructCreateATAInstruction(userPublicKey, tokenMint)
	if err != nil {
		t.Errorf("Test failed, expected: %v, got: %v", nil, err)
	}

}

func Test_GetSellPercentage(t *testing.T) {
	sellPercentage, err := GetSellPercentage(100*1e6, 10*1e6)
	if err != nil {
		t.Error(err)
	}
	if sellPercentage != 0.9 {
		t.Errorf("Test failed, expected: %v, got: %v", 0.9, sellPercentage)
	}
}

func Test_GetSellPercentage2(t *testing.T) {

	sellPercentage, err := GetSellPercentage(100*1e6, 50*1e6)
	if err != nil {
		t.Error(err)
	}
	if sellPercentage != 0.5 {
		t.Errorf("Test failed, expected: %v, got: %v", 0.5, sellPercentage)
	}
}

var tableBenchmarkConstructCreateATAInstruction = []struct {
	userPublicKey string
	tokenMint     string
}{
	{
		userPublicKey: "ZDK6SbnjEtgTWG1CqHhNueEFvziCcMQYQhW2UwJ3EzU",
		tokenMint:     "EeJ1TH7MACdR8dmHBaGGrkZrsHqWG86J74F9m5fZpump",
	},
}

func BenchmarkConstructCreateATAInstruction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, table := range tableBenchmarkConstructCreateATAInstruction {
			_, err := utils.ConstructCreateATAInstruction(table.userPublicKey, table.tokenMint)
			if err != nil {
				return
			}
		}
	}
}
