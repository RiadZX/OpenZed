package triggers

import "github.com/gookit/slog"

// StopLossTrigger triggers when the price falls below the stop-loss level
type StopLossTrigger struct{}

// ShouldTrigger implements the Trigger interface for StopLossTrigger
func (t *StopLossTrigger) ShouldTrigger(trade TradeInterface, opts *RPCOpts) (bool, error) {
	price, err := trade.GetPrice(opts)
	if err != nil {
		return false, err
	}
	pnl := trade.CalculateProfitLoss(price)
	slog.Info("StopLossTrigger for mint: ", trade.GetMint(), " PnL: ", pnl)
	if pnl <= trade.GetStopLoss() {
		return true, nil
	}
	return false, nil
}
