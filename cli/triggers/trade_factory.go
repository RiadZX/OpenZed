package triggers

import (
	"Zed/pumpfunsdk"
	"Zed/raydiumsdk/ray_idl"
	"github.com/gagliardetto/solana-go"
	"math/rand"
)

// PumpfunTradeFactory implements TradeFactory
type PumpfunTradeFactory struct{}

// CreateTrade creates a new trade based on the platform
func (f *PumpfunTradeFactory) CreateTrade(mint string, entryPrice float64, stoploss float64, takeprofit float64, opts ExecutionOpts) TradeInterface {
	randomId := rand.Intn(1000000)

	return &PumpfunTrade{
		Trade: Trade{
			ID:            randomId,
			Mint:          mint,
			EntryPrice:    entryPrice,
			StopLoss:      stoploss,
			TakeProfit:    takeprofit,
			active:        true,
			ExecutionOpts: opts,
		},
		BondingCurve: pumpfunsdk.BondingCurveFromMint(solana.MustPublicKeyFromBase58(mint)),
	}

}

// RaydiumTradeFactory implements TradeFactory
type RaydiumTradeFactory struct{}

func (f *RaydiumTradeFactory) CreateTrade(mint string, entryPrice float64, stoploss float64, takeprofit float64, decimals int, in ray_idl.SwapBaseIn, opts ExecutionOpts) TradeInterface {
	randomId := rand.Intn(1000000)

	return &RaydiumTrade{
		Trade: Trade{
			ID:            randomId,
			Mint:          mint,
			EntryPrice:    entryPrice,
			StopLoss:      stoploss,
			TakeProfit:    takeprofit,
			active:        true,
			ExecutionOpts: opts,
		},
		// ... initialize Raydium-specific fields ...
		decimals:        decimals,
		swapInstruction: in,
	}

}
