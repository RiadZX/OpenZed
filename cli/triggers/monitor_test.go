package triggers

import (
	"Zed/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMonitor(t *testing.T) {
	opts := RPCOpts{}
	err := NewMonitor(&models.Config{}, 100, opts)
	assert.NoError(t, err, "Monitor creation should not return an error")

	monitorImpl, ok := MonitorInstance.(*MonitorImpl)
	assert.True(t, ok, "MonitorInstance should be of type MonitorImpl")
	assert.Equal(t, 2, len(monitorImpl.triggers), "Monitor should have 2 triggers")
	assert.Equal(t, 100, monitorImpl.delay, "Monitor should have a delay of 100")
	assert.Equal(t, opts, monitorImpl.RPCOpts, "Monitor should have the same RPCOpts")

}

func TestAddTrade(t *testing.T) {
	opts := RPCOpts{}
	err := NewMonitor(&models.Config{}, 100, opts)
	assert.NoError(t, err, "Monitor creation should not return an error")

	fac := PumpfunTradeFactory{}
	trade := fac.CreateTrade("AS5UzBvxkXPuKiAYsA4R8G6QY16jSVqitxR4hjaypump", 1, 0.9, 1.2, ExecutionOpts{})
	err = MonitorInstance.AddTrade(trade)
	assert.NoError(t, err, "Adding trade should not return an error")

	monitorImpl, ok := MonitorInstance.(*MonitorImpl)
	assert.True(t, ok, "MonitorInstance should be of type MonitorImpl")
	assert.Equal(t, 1, len(monitorImpl.trades), "Monitor should have 1 trade")
}
