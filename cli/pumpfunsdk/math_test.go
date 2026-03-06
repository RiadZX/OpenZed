package pumpfunsdk

import (
	"testing"
)

func TestCalculateTokensToBuy(t *testing.T) {
	// Create a new BondingCurve with some values.
	bondingCurve := &BondingCurve{
		VirtualTokenReserves: 1071214586840484,
		VirtualSolReserves:   30050001791,
		RealTokenReserves:    791314586840484,
		RealSolReserves:      50001791,
		Complete:             false,
	}
	// Calculate the amount of tokens that can be bought with 1.5 sol.
	amountOfSol := 1.0
	tokens := CalculateTokensToBuy(bondingCurve, amountOfSol)
	if tokens != 34499662642563 {
		t.Errorf("Expected 34499662642563, got %d", tokens)
	}
	// other case
	amountOfSol = 2100
	tokens = CalculateTokensToBuy(bondingCurve, amountOfSol)
	if tokens != 791314586840484 {
		t.Errorf("Expected 6895 * 1e6, got %d", tokens)
	}

}

func TestCalculateTokenPrice(t *testing.T) {
	// Create a new BondingCurve with some values.
	bondingCurve := &BondingCurve{
		VirtualTokenReserves: 1072999997516496,
		VirtualSolReserves:   30000000073,
		RealTokenReserves:    793099997516496,
		RealSolReserves:      73,
		TokenTotalSupply:     1000000000000000,
		Complete:             false,
	}
	// Calculate the price of a token.
	price, err := CalculateTokenPrice(bondingCurve)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if price != 27.958993608980684 {
		t.Errorf("Expected 27.958993608980684, got %f", price)
	}

	bondingCurve.Complete = true
	price, err = CalculateTokenPrice(bondingCurve)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	bondingCurve.Complete = false
	bondingCurve.VirtualTokenReserves = 0
	price, err = CalculateTokenPrice(bondingCurve)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	bondingCurve.VirtualTokenReserves = 1072999997516496
	bondingCurve.VirtualSolReserves = 0
	price, err = CalculateTokenPrice(bondingCurve)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
