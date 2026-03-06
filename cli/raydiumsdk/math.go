package raydiumsdk

import (
	"Zed/utils"
	"fmt"
	"math"
)

// GetSellPercentage tokensSold is the amount of tokens the walelt has told, this is in decimals so already multiplied by 1e6
func GetSellPercentage(preTokenBalance, postTokenBalance float64) (float64, error) {
	if postTokenBalance == 0 {
		return 1, nil
	}

	percentage := (preTokenBalance - postTokenBalance) / preTokenBalance

	if percentage > 1.01 {
		return 0, fmt.Errorf("math error")
	}
	return percentage, nil
}

// CalculatePricePerToken calculates the price per token in lamports.
func CalculatePricePerToken(balances *utils.CustomPrePostTokenBalances, mint string) (pricePerToken uint64, err error) {
	preTokenBalance := balances.GetPreTokenBalance(rayAuthority.String(), mint)
	postTokenBalance := balances.GetPostTokenBalance(rayAuthority.String(), mint)
	preSolBalance := balances.GetPreTokenBalance(rayAuthority.String(), Wsol.String())
	postSolBalance := balances.GetPostTokenBalance(rayAuthority.String(), Wsol.String())

	// Get the decimals for the token and SOL
	tokenDecimals := GetDecimals(balances, mint)

	if tokenDecimals == nil {
		return 0, fmt.Errorf("could not retrieve decimals for token or SOL")
	}

	// Adjust balances by their respective decimals
	preTokenBalance /= math.Pow(10, float64(*tokenDecimals))
	postTokenBalance /= math.Pow(10, float64(*tokenDecimals))

	if postTokenBalance == 0 {
		return 0, fmt.Errorf("postTokenBalance is 0")
	}
	if postSolBalance == 0 {
		return 0, fmt.Errorf("postSolBalance is 0")
	}

	// Calculate the price per token in lamports.
	pricePerToken = uint64(math.Abs((postSolBalance - preSolBalance) / (postTokenBalance - preTokenBalance)))
	return pricePerToken, nil
}
