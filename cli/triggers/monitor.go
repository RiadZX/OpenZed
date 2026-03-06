package triggers

import (
	"Zed/models"
	"errors"
	"github.com/gookit/slog"
	"time"
)

// Monitor Is The Main Part of this module
// It is responsible for keeping track of the trades, and receiving the new trades.
// It also keeps track of the active triggers, and triggers them when necessary.
// It is the main entry point for the module.
// It's responsibilities are:
// - Keep track of the trades
// - Keep track of the triggers
// - Receive new trades
// - Trigger the triggers when necessary
// - Get Price of All Trades and Check if any trigger is triggered, every x seconds defined by the user.
// - Handle the logic of the triggers

var (
	MonitorInstance Monitor
)

// Monitor defines the interface for trade triggers
type Monitor interface {
	// AddTrade adds a new trade to the monitor
	AddTrade(trade TradeInterface) error
	// AddTrigger adds a new trigger to the monitor
	AddTrigger(trigger Trigger) error
	// Start starts the monitoring process
	Start()
	Stop()
}

// MonitorImpl implements Monitor
type MonitorImpl struct {
	trades     []TradeInterface
	triggers   []Trigger
	delay      int // in Milliseconds
	userConfig *models.Config
	// field to handle stopping the monitor
	stop    bool
	RPCOpts RPCOpts
}

func (m *MonitorImpl) Stop() {
	// Stop the monitor
	m.stop = true
}

func NewMonitor(conf *models.Config, PnlDelayMS int, opts RPCOpts) error {
	m := MonitorImpl{
		delay:      PnlDelayMS,
		RPCOpts:    opts,
		userConfig: conf,
	}
	err := m.AddTrigger(&StopLossTrigger{})
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	err = m.AddTrigger(&TakeProfitTrigger{})
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	MonitorInstance = &m
	return nil
}

func (m *MonitorImpl) AddTrade(trade TradeInterface) error {
	if trade == nil {
		return errors.New("trade cannot be nil")
	}
	if m.trades == nil {
		m.trades = make([]TradeInterface, 0)
	}
	// check if the trade with the same signature already exists
	for _, t := range m.trades {
		if t.GetExecutionOpts().Signature == trade.GetExecutionOpts().Signature {
			return nil
		}
	}
	m.trades = append(m.trades, trade)
	return nil
}

func (m *MonitorImpl) AddTrigger(trigger Trigger) error {
	if trigger == nil {
		return errors.New("trigger cannot be nil")
	}
	if m.triggers == nil {
		m.triggers = make([]Trigger, 0)
	}
	m.triggers = append(m.triggers, trigger)
	return nil
}

// Start starts the monitoring process.
// This will live forever, and will keep checking the price of all trades every `MonitorImpl.delay` milliseconds
// If any trigger is triggered, it will execute the trigger.
// If the trigger is successful, it will remove the trade from the list of trades.
// If the trigger is unsuccessful, it will keep the trade in the list of trades.
// This will keep running until the program is stopped.
func (m *MonitorImpl) Start() {
	m.stop = false
	go func() {
		for {
			if m.stop {
				return
			}
			for _, trade := range m.trades {
				//slog.Debug("Checking trade: " + spew.Sdump(trade))
				if !trade.Active() {
					continue
				}
				for _, trigger := range m.triggers {
					shouldTrigger, err := trigger.ShouldTrigger(trade, &m.RPCOpts)
					if err != nil {
						if err.Error() == "bonding_curve_is_complete" {
							slog.Info("Bonding curve is complete. Removing trade.")
							trade.SetActive(false)
						} else {
							slog.Error("Error checking trigger: ", err.Error())
						}

					}
					if shouldTrigger {
						switch trigger.(type) {
						case *StopLossTrigger:
							err := trade.ExecuteSell(m.userConfig)
							if err != nil {
								slog.Error("Error executing stop-loss order: ", err.Error())
								return
							}
							slog.Info("Stop-loss order executed")
							// Remove the trade from the list of trades
							return
						case *TakeProfitTrigger:
							err := trade.ExecuteSell(m.userConfig)
							if err != nil {
								return
							}
							slog.Info("Take-profit order executed")
							return
						default:
							slog.Error("Unknown trigger type")
						}
					}
				}
			}

			// Sleep for the delay
			time.Sleep(time.Duration(m.delay) * time.Millisecond)
		}
	}()

}
