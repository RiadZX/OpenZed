package triggers

// Trigger defines the interface for trade triggers
type Trigger interface {
	ShouldTrigger(trade TradeInterface, opts *RPCOpts) (bool, error)
}
