package raydiumsdk

import "github.com/gagliardetto/solana-go/rpc"

// IsWSOLSwap returns true if it is a SOL/X or X/SOL Swap
func IsWSOLSwap(tokenBalances []rpc.TokenBalance) (valid bool) {
	if len(tokenBalances) == 0 {
		return false
	}

	for _, postTokenBalance := range tokenBalances {
		if postTokenBalance.Mint == Usdc {
			//slog.Info("USDC SWAP DETECTED " + tx.Signatures[0].String())
			return false
		}
		if postTokenBalance.Mint == Wsol {
			return true
		}
	}
	return false
}
