package triggers

import (
	"Zed/models"
	strategy "Zed/strategy/transaction"
	"github.com/gagliardetto/solana-go"
)

type Trade struct {
	ID         int
	active     bool
	Mint       string
	EntryPrice float64
	StopLoss   float64
	TakeProfit float64
	ExecutionOpts
}

type ExecutionOpts struct {
	Signature  solana.Signature
	Wallet     solana.Wallet
	FixedFee   float64
	Sender     strategy.SendStrategy
	TokenCache *models.TokensOwnedCache
	Blockhash  *solana.Hash
}

type TradeInterface interface {
	Active() bool
	// CalculateProfitLoss calculates the profit or loss based on the current price.
	CalculateProfitLoss(currentPrice float64) float64
	GetStopLoss() float64
	GetTakeProfit() float64
	// GetPrice returns the current price of the mint.
	GetPrice(opts *RPCOpts) (float64, error)
	GetMint() string

	GetExecutionOpts() ExecutionOpts
	// ExecuteSell sells the token.
	ExecuteSell(config *models.Config) error
	SetActive(active bool)
}

// GetStopLoss returns the stop loss of the trade
func (t *Trade) GetStopLoss() float64 {
	return t.StopLoss
}

// GetTakeProfit returns the take profit of the trade
func (t *Trade) GetTakeProfit() float64 {
	return t.TakeProfit
}
