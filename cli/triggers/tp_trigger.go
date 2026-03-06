package triggers

import "github.com/gookit/slog"

// TakeProfitTrigger triggers when the price goes above the take-profit level
type TakeProfitTrigger struct{}

// ShouldTrigger implements the Trigger interface for TakeProfitTrigger
func (t *TakeProfitTrigger) ShouldTrigger(trade TradeInterface, opts *RPCOpts) (bool, error) {
	price, err := trade.GetPrice(opts)
	if err != nil {
		slog.Error(err.Error())
		return false, err
	}
	pnl := trade.CalculateProfitLoss(price)
	slog.Info("TakeProfitTrigger for mint: ", trade.GetMint(), " PnL: ", pnl)
	if pnl >= trade.GetTakeProfit() {
		return true, nil
	}
	return false, nil
}
