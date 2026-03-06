package raydiumsdk

import (
	"Zed/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculatePricePerToken_ZeroPostTokenBalance(t *testing.T) {
	balances := utils.CustomPrePostTokenBalances{
		PreTokenBalances: []utils.XTokenBalance{
			{Mint: "token1", Owner: "owner1", Tokenites: 100000000, Decimals: 6},
			{Mint: Wsol.String(), Owner: "owner1", Tokenites: 200000000000, Decimals: 9},
		},
		PostTokenBalances: []utils.XTokenBalance{
			{Mint: "token1", Owner: "owner1", Tokenites: 0, Decimals: 6},
			{Mint: Wsol.String(), Owner: "owner1", Tokenites: 250000000000, Decimals: 9},
		},
	}

	pricePerToken, err := CalculatePricePerToken(&balances, "token1")
	assert.Error(t, err, "CalculatePricePerToken should return an error when postTokenBalance is 0")
	assert.Equal(t, uint64(0), pricePerToken, "Price per token should be 0 when postTokenBalance is 0")
}
