package triggers

import (
	"Zed/models"
	"Zed/pumpfunsdk"
	strategy "Zed/strategy/transaction"
	"Zed/utils"
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
)

type PumpfunTrade struct {
	Trade
	// This is the bonding curve at the moment of the trade was made. i.e. the entry bonding curve.
	BondingCurve solana.PublicKey
}

func (t *PumpfunTrade) Active() bool {
	return t.active
}

// CalculateProfitLoss for PumpfunTrade
func (t *PumpfunTrade) CalculateProfitLoss(currentPrice float64) float64 {
	if t.EntryPrice == 0 {
		return 10 // 10 is a placeholder value
	}
	// Calculate the profit or loss based on the current price.
	// if difference is 50% -> 1.5
	// if -50% -> 0.5
	return currentPrice / t.EntryPrice
}

// GetPrice returns the current price of the mint.
func (t *PumpfunTrade) GetPrice(opts *RPCOpts) (float64, error) {
	// Use the variables in the bonding curve to get the current price.
	var bondingCurve *pumpfunsdk.BondingCurve
	var err error
	bondingCurve, err = pumpfunsdk.GetBondingCurve(opts.Client, t.BondingCurve, pumpfunsdk.ClientOpts{
		Commitment: rpc.CommitmentProcessed,
		Retries:    opts.Retries,
		Delay:      opts.DelayMS,
	})
	if err != nil {
		slog.Error("Error getting bonding curve:", err)
		return -1, err
	}
	pricePerTokens, err := pumpfunsdk.CalculateTokenPrice(bondingCurve)
	if err != nil {
		slog.Error("Error: ", err)
		return -1, err
	}
	if bondingCurve.Complete {
		slog.Info("Bonding curve is complete.")
		return -1, errors.New("bonding_curve_is_complete")
	}
	return pricePerTokens, nil
}

func (t *PumpfunTrade) GetExecutionOpts() ExecutionOpts {
	return t.ExecutionOpts
}

func (t *PumpfunTrade) GetMint() string {
	return t.Mint
}

// SetActive sets the active status of the trade
func (t *PumpfunTrade) SetActive(active bool) {
	t.active = active
}

func (t *PumpfunTrade) ExecuteSell(config *models.Config) error {
	slog.Info("Executing sell for PumpfunTrade with mint: ", t.Mint)
	// Sell the token
	associatedUser, _, err := solana.FindAssociatedTokenAddress(t.Wallet.PublicKey(), solana.MustPublicKeyFromBase58(t.Mint))
	if err != nil {
		slog.Error("Error finding associated token address", err)
		return err
	}
	err = t.TokenCache.Preload(config.HttpClient, t.Wallet.PublicKey().String())
	if err != nil {
		slog.Error("Error preloading token cache: ", err)
		return err
	}
	if !t.TokenCache.Exists(t.Mint) {
		return errors.New("token not owned")
	}
	tokenAmount, exists := t.TokenCache.Get(t.Mint)
	if !exists {
		return errors.New("token not owned")
	}

	var sellIX solana.Instruction
	associatedBondingCurve := pumpfunsdk.AssociatedBondingCurveFromMint(solana.MustPublicKeyFromBase58(t.Mint))
	sellIX, err = pumpfunsdk.ConstructSellInstruction(
		uint64(tokenAmount*1e6),
		0,
		solana.MustPublicKeyFromBase58(t.Mint),
		t.BondingCurve,
		associatedBondingCurve,
		associatedUser,
		t.Wallet.PublicKey(),
	)
	if err != nil {
		slog.Error("Error: ", err)
		return err
	}
	var instructions = []solana.Instruction{sellIX}
	computeUnitLimitIx, _ := utils.ComputeBudgetLimit(130_000)
	computeBudgetPriceIx, _ := utils.ComputeBudgetPrice(uint64(t.FixedFee * 1e9))
	instructions = append([]solana.Instruction{computeBudgetPriceIx.Build(), computeUnitLimitIx.Build()}, instructions...)
	tipIx, err := t.Sender.GetTipInstructions(t.Wallet, config.CopyTraderConfig.SellTipFee)
	if err != nil {
		slog.Error("Error getting tip instructions: ", err)
		return err
	}
	instructions = append(instructions, tipIx...)
	tx, err := solana.NewTransaction(instructions, *t.Blockhash, solana.TransactionPayer(t.Wallet.PublicKey()))
	if err != nil {
		slog.Error("Error creating transaction", err)
		return err
	}

	// sign the transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &t.Wallet.PrivateKey
	})
	if err != nil {
		slog.Error("Error signing transaction", err)
		return err
	}
	// Send the transaction
	ok, err := t.Sender.SendTransaction(*tx, strategy.SendParams{
		TaskName:         "PumpFun RiskManager",
		SimulationFailed: false,
	})
	if err != nil {
		slog.Error("Error sending transaction: ", err)
		return err
	}
	if ok {
		slog.Info("Successfully sent transaction: " + tx.Signatures[0].String())
	} else {
		slog.Warn("Could not confirm transaction")
	}
	// --- snip ---
	t.active = false
	return nil
}
