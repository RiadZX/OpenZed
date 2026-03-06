package raydiumsdk

import (
	"Zed/utils"
	"github.com/gagliardetto/solana-go"
	"math"
)

// GetDecimals returns the decimals of a token. Returns nil if the token is not found.
func GetDecimals(pptb *utils.CustomPrePostTokenBalances, token string) *uint8 {
	preTokenBalances := pptb.PreTokenBalances
	postTokenBalances := pptb.PostTokenBalances
	for _, balance := range preTokenBalances {
		if balance.Mint == token {
			return &balance.Decimals
		}
	}
	for _, balance := range postTokenBalances {
		if balance.Mint == token {
			return &balance.Decimals
		}
	}
	return nil
}

// GetSupply returns the supply of a token based on the isBuy flag. Returns nil if the token is not found.
// Returns the amount in full units (not in tokenites)
// 1,2 ,3 not 1000000000, 2000000000, 3000000000
func GetSupply(pptb *utils.CustomPrePostTokenBalances, isBuy bool, token string) *float64 {
	preTokenBalancesMap := make(map[uint16]utils.XTokenBalance)
	postTokenBalancesMap := make(map[uint16]utils.XTokenBalance)
	preTokenBalances := pptb.PreTokenBalances
	postTokenBalances := pptb.PostTokenBalances
	for _, balance := range preTokenBalances {
		preTokenBalancesMap[balance.AccountIndex] = balance
	}
	for _, balance := range postTokenBalances {
		postTokenBalancesMap[balance.AccountIndex] = balance
	}

	for accountIndex, preBalance := range preTokenBalancesMap {
		if preBalance.Owner == rayAuthority.String() && preBalance.Mint == token {
			postBalance, exists := postTokenBalancesMap[accountIndex]
			if exists {
				// Process it with the decimals
				preAmount := preBalance.Tokenites / math.Pow(10, float64(preBalance.Decimals))
				postAmount := postBalance.Tokenites / math.Pow(10, float64(postBalance.Decimals))
				if isBuy {
					return &postAmount
				} else {
					return &preAmount
				}
			}
		}
	}
	return nil
}

// GetMintFromBalances returns the mint of the token and whether it's a buy or not.
func GetMintFromBalances(pptb *utils.CustomPrePostTokenBalances, signer solana.PublicKey) (bool, string) {
	preTokenBalancesMap := make(map[uint16]utils.XTokenBalance)
	preTokenBalances := pptb.PreTokenBalances
	postTokenBalances := pptb.PostTokenBalances
	for _, balance := range preTokenBalances {
		preTokenBalancesMap[balance.AccountIndex] = balance
	}

	var isBuy bool
	var token string

	for _, postBalance := range postTokenBalances {
		accountIndex := postBalance.AccountIndex
		preBalance, exists := preTokenBalancesMap[accountIndex]

		if exists {
			owner := preBalance.Owner
			mint := preBalance.Mint
			preAmount := preBalance.Tokenites
			postAmount := postBalance.Tokenites
			amountDifference := preAmount - postAmount

			if owner == rayAuthority.String() && mint == Wsol.String() {
				if amountDifference > 0 {
					isBuy = false
				} else if amountDifference < 0 {
					isBuy = true
				}
			} else if owner == signer.String() && mint != Wsol.String() {
				token = mint
				if amountDifference > 0 {
					isBuy = false
				} else if amountDifference < 0 {
					isBuy = true
				}
			}
		} else {
			// New token account balance created
			if postBalance.Mint != Wsol.String() {
				token = postBalance.Mint
				postAmount := postBalance.Tokenites
				if postAmount > 0 {
					isBuy = true
				}
			}
		}
	}

	return isBuy, token
}
