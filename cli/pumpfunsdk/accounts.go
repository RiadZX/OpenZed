package pumpfunsdk

import "github.com/gagliardetto/solana-go"

type BondingCurve struct {
	VirtualTokenReserves uint64
	VirtualSolReserves   uint64
	RealTokenReserves    uint64
	RealSolReserves      uint64
	TokenTotalSupply     uint64
	Complete             bool
}

func BondingCurveFromMint(mint solana.PublicKey) solana.PublicKey {
	p, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("bonding-curve"),
			mint[:],
		},
		PUMP_PROGRAM_ADDRESS,
	)
	if err != nil {
		panic(err)
	}
	return p
}

func AssociatedBondingCurveFromMint(mint solana.PublicKey) solana.PublicKey {
	bondingCurve := BondingCurveFromMint(mint)
	p, _, err := solana.FindAssociatedTokenAddress(bondingCurve, mint)
	if err != nil {
		panic(err)
	}
	return p
}
