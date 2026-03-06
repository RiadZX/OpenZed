package utils

import (
	"Zed/models"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gookit/slog"
)

// GetOptimalVersionedTransaction sends the transaction in the most optimal way
// this will be used to send any compiled versioned transaction.
// this might also handle the retries, jito, bloxroute, optimal fees, optimal Compute Units etc.
// Can return ErrSimulationFailed if the simulation fails, in that case use normal transaction sending, not JITO.
func GetOptimalVersionedTransaction(
	client models.IClient, ct models.CopyTraderConfig, blockHash *solana.Hash, buy bool, task *models.CopyTraderTask, program string, instructions ...solana.Instruction) (solana.Transaction, error) {
	if client == nil {
		return solana.Transaction{}, errors.New("client is nil")
	}
	if task == nil {
		return solana.Transaction{}, errors.New("task is nil")
	}
	if task.Wallet == nil {
		return solana.Transaction{}, errors.New("wallet  is nil")
	}
	if blockHash == nil {
		return solana.Transaction{}, errors.New("blockHash is nil")
	}

	var err error
	slog.Debug("Entering SendOptimalVersionedTransaction")

	//get the priority fee estimate from an api. give option from helius or quicknode or solana?? just an api call.
	//priorFee := txConf.PriorityFee
	var priorFee uint64
	if task.DynamicFee {
		priorFee, err = GetPriorityFeeEstimate(ct.HeliusUrl, task.DynamicFeeLevel, program)
		if err != nil {
			slog.Error("Error getting priority fee estimate: " + err.Error())
			priorFee = 5e6
			slog.Info("Using default priority fee: 5e6")
		}
	} else {
		priorFee = uint64(task.FixedFee * 1e9)
	}

	computeBudgetPriceIX, err := ComputeBudgetPrice(priorFee)
	if err != nil {
		return solana.Transaction{}, err
	}
	slog.Debug("Got the priority fee estimate")
	//get the unit limit estimate

	var unitLimit uint32 = 130_000
	var errMain error = nil

	//unitLimit, errMain = GetUnitLimitEstimate(client, task.Wallet, blockHash, instructions)
	//if errMain != nil || unitLimit == 0 {
	//	//Since we can't get the unit limit, we will use a hardcoded value
	//	if errMain != nil {
	//		slog.Debug("Error getting unit limit estimate: " + errMain.Error())
	//	}
	//	slog.Debug("Using hardcoded value")
	//	unitLimit = 110_000
	//}

	unitLimit = uint32(float64(unitLimit) * 1.1)
	slog.Debug("Unit limit after calc: ", unitLimit)

	setComputeLimitIX, err := ComputeBudgetLimit(unitLimit)
	if err != nil {
		return solana.Transaction{}, err
	}
	slog.Debug("Set the compute limit")
	//build the transaction using these instructions
	var tx *solana.Transaction
	instructions = append([]solana.Instruction{computeBudgetPriceIX.Build(), setComputeLimitIX.Build()}, instructions...)
	tipFee := ct.BuyTipFee
	if !buy {
		tipFee = ct.SellTipFee
	}
	tipInstructions, err := ct.Sender.GetTipInstructions(*task.Wallet, tipFee)
	if err != nil {
		return solana.Transaction{}, err
	}
	instructions = append(instructions, tipInstructions...)

	if len(instructions) > 0 {
		tx, err = solana.NewTransaction(instructions, *blockHash, solana.TransactionPayer(task.Wallet.PublicKey()))
		if err != nil {
			return solana.Transaction{}, err
		}
	} else {
		return solana.Transaction{}, errors.New("no instructions provided")
	}
	slog.Debug("Built the transaction")

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &task.Wallet.PrivateKey
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return solana.Transaction{}, err
	}
	slog.Debug("Signed the transaction")

	// return the transaction and the error
	// It returns the error related to simulation, if the simulation fails, use normal sending.
	return *tx, errMain
}
