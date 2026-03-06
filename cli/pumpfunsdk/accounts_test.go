package pumpfunsdk

import (
	"github.com/gagliardetto/solana-go"
	"testing"
)

func Test_AssociatedBondingCurveFromMint(t *testing.T) {
	mint := solana.MustPublicKeyFromBase58("AS5UzBvxkXPuKiAYsA4R8G6QY16jSVqitxR4hjaypump")
	bondingCurve := BondingCurveFromMint(mint)
	assBondingCurve := AssociatedBondingCurveFromMint(mint)

	if bondingCurve.String() != "9hE8cSF3fhqyWZLxD3nXv9JgBtYwek3Tbu4wzfg3Woxd" {
		t.Errorf("unexpected")
	}
	if assBondingCurve.String() != "CWhCcRk4JNQejbJeznQ2iNA7SfCFF3cboZDPGA4YRtEz" {
		t.Errorf("unexpected")
	}
}
