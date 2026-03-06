package pumpfunsdk

import (
	"errors"
	"math"
	"math/big"
)

// CalculateTokensToBuy
// From the BondingCurve and the amount of sol, calculate the amount of tokens that can be bought.
//
// input is the amount of sol in float64 1.5 sol = 1.5
// output is the amount of tokens multiplied by the decimals. so this can directly be used in the transaction
func CalculateTokensToBuy(bondingCurve *BondingCurve, amountOfSol float64) (tokenites uint64) {
	if bondingCurve.Complete {
		return 0
	}
	virtualSOLReservesBN := big.NewInt(int64(bondingCurve.VirtualSolReserves))
	virtualTokenReservesBN := big.NewInt(int64(bondingCurve.VirtualTokenReserves))
	reservesProduct := new(big.Int).Mul(virtualSOLReservesBN, virtualTokenReservesBN)
	newVirtualSOLReserve := new(big.Int).Add(virtualSOLReservesBN, big.NewInt(int64(amountOfSol*1e9)))
	newVirtualTokenReserve := new(big.Int).Div(reservesProduct, newVirtualSOLReserve)
	newVirtualTokenReserve = new(big.Int).Add(newVirtualTokenReserve, big.NewInt(1))
	amountOut := new(big.Int).Sub(virtualTokenReservesBN, newVirtualTokenReserve)
	finalAmountOut := amountOut.Uint64()

	if amountOut.Uint64() > bondingCurve.RealTokenReserves {
		finalAmountOut = bondingCurve.RealTokenReserves //if we dont have enough tokens in the reserve, buy the remaining reserve
	}

	return finalAmountOut
}

// CalculateTokenPrice
// From the BondingCurve, calculate the price of a token. Returns the amount of lamports you need to pay for 1 full token. (so not 1 decimal, rather 1 full token)
//
// No need to multiply by 1e9, this is already in lamports.
func CalculateTokenPrice(bondingCurve *BondingCurve) (float64, error) {
	if bondingCurve.Complete {
		return 0, errors.New("bonding curve is complete")
	}
	if bondingCurve.VirtualTokenReserves == 0 || bondingCurve.VirtualSolReserves == 0 {
		return 0, errors.New("bonding curve has no reserves")
	}
	// Calculate the price of a token in lamports.
	adjustedVTR := float64(bondingCurve.VirtualTokenReserves) / math.Pow(10, 6)
	adjustedVSR := float64(bondingCurve.VirtualSolReserves)

	return adjustedVSR / adjustedVTR, nil
}
