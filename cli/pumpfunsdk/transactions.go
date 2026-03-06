package pumpfunsdk

import (
	"Zed/pumpfunsdk/idl"
	"errors"
	"github.com/gagliardetto/solana-go"
)

func ConstructBuyInstruction(
	userPublicKey solana.PublicKey,
	tokenAmount uint64,
	maxLamportAmount uint64,
	tokenMint solana.PublicKey,
	bondingCurve solana.PublicKey,
	bondingCurveAssociated solana.PublicKey,
	associatedUserTokenAccount solana.PublicKey,
) (solana.Instruction, error) {
	if tokenAmount == 0 || maxLamportAmount == 0 {
		return nil, errors.New("tokenAmount and maxLamportAmount must be greater than 0")
	}

	buyIx := pump_idl.NewBuyInstruction(
		tokenAmount,
		maxLamportAmount,
		GLOBAL_ACCOUNT,
		FEE_RECIPIENT,
		tokenMint,
		bondingCurve,
		bondingCurveAssociated,
		associatedUserTokenAccount,
		userPublicKey,
		solana.SystemProgramID,
		solana.TokenProgramID,
		solana.SysVarRentPubkey,
		EVENT_AUTHORITY,
		PUMP_PROGRAM_ADDRESS,
	)

	err := buyIx.Validate()
	if err != nil {
		return nil, err
	}

	builtIx := buyIx.Build()

	//convert to solana.Instruction
	data, err := builtIx.Data()
	if err != nil {
		return nil, err
	}
	solanaIx := solana.NewInstruction(
		builtIx.ProgramID(),
		builtIx.Accounts(),
		data,
	)

	return solanaIx, nil
}

// ConstructSellInstruction constructs a sell transaction.
// The amount is the amount of tokens to sell accounting the decimals, so you should multiply the amount by 10^decimals.
// The minSolOutput is the minimum amount of SOL you want to receive in lamports.
func ConstructSellInstruction(
	amount uint64,
	minSolOutput uint64,
	mint solana.PublicKey,
	bondingCurve solana.PublicKey,
	associatedBondingCurve solana.PublicKey,
	ata solana.PublicKey,
	user solana.PublicKey,
) (solana.Instruction, error) {
	if amount == 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	sellIx := pump_idl.NewSellInstruction(
		amount,
		minSolOutput,
		GLOBAL_ACCOUNT,
		FEE_RECIPIENT,
		mint,
		bondingCurve,
		associatedBondingCurve,
		ata,
		user,
		solana.SystemProgramID,
		ATA_TOKEN_PROGRAM,
		solana.TokenProgramID,
		EVENT_AUTHORITY,
		PUMP_PROGRAM_ADDRESS)

	err := sellIx.Validate()
	if err != nil {
		return nil, err
	}

	builtIx := sellIx.Build()
	return builtIx, nil
}
