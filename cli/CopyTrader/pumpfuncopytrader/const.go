package pumpfuncopytrader

import (
	"errors"
)

var ErrConfigNotAvailable = errors.New("config not available")
var ErrNotInitialized = errors.New("copyTrader not initialized")
