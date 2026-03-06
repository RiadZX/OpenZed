package triggers

import (
	"Zed/models"
	"Zed/raydiumsdk"
	"Zed/raydiumsdk/ray_idl"
	strategy "Zed/strategy/transaction"
	"Zed/utils"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gookit/slog"
	"math"
)

type RaydiumTrade struct {
	Trade
	decimals        int
	swapInstruction ray_idl.SwapBaseIn
}

func (r *RaydiumTrade) ExecuteSell(config *models.Config) error {
	//Amount in is in token decimals -> Get this from the tokens owned cache, since we need to know the amount of tokens owned. -> if not in cache then we can check if it exists
	if !r.TokenCache.Exists(r.Mint) {
		//Fetch the tokens owned
		err := r.TokenCache.Preload(config.HttpClient, r.Wallet.PublicKey().String())
		if err != nil {
			slog.Error("error: " + err.Error())
			return err
		}
		if !r.TokenCache.Exists(r.Mint) {
			slog.Error("Token not owned: " + r.Mint)
			return nil
		}
	}

	tokenAmount, exists := r.TokenCache.Get(r.Mint) //tokens owned, this is 100%, this is the ui amount
	if !exists {
		slog.Error("Token not owned: " + r.Mint)
		return nil
	}
	if tokenAmount == 0 {
		slog.Info("You have no tokens to sell :/")
		return nil
	}
	//direction is copy-pasted from the eventData
	//Instructions needed
	//Create account with seed for wsol
	//initialize the seeded wsol
	//swap ix
	//close the seeded sol.
	var createATAWSOLSEED *system.Instruction
	var initSeedWSOL *token.Instruction
	var closeWSOLATASEED *token.Instruction
	var swapIx solana.Instruction
	//creating the ata seed account
	seed := solana.NewWallet().PublicKey().String()[0:32]
	//concat the public key + seed + program id into sha256 to create a new public key
	tokenProgram := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	wsolAccount, err := solana.CreateWithSeed(r.Wallet.PublicKey(), seed, tokenProgram)
	createATAWSOLSEED, err = system.NewCreateAccountWithSeedInstruction(
		r.Wallet.PublicKey(),
		seed,
		3e6,
		165,
		tokenProgram,
		r.Wallet.PublicKey(),
		wsolAccount,
		r.Wallet.PublicKey(),
	).ValidateAndBuild()
	if err != nil {
		slog.Error("error: " + err.Error())
		return err
	}
	//init the seeded ata
	initSeedWSOL, err = token.NewInitializeAccountInstruction(wsolAccount, raydiumsdk.Wsol, r.Wallet.PublicKey(), solana.SysVarRentPubkey).ValidateAndBuild()
	if err != nil {
		slog.Error("error: " + err.Error())
		return err
	}
	//close the wsol seed ata ix
	//close the seed account
	closeWSOLATASEED, err = token.NewCloseAccountInstruction(
		wsolAccount,
		r.Wallet.PublicKey(),
		r.Wallet.PublicKey(),
		[]solana.PublicKey{r.Wallet.PublicKey()},
	).ValidateAndBuild()
	if err != nil {
		slog.Error("error: " + err.Error())
		return err
	}
	mintTokenAccount, err := utils.GetATA(r.Wallet.PublicKey().String(), r.Mint)
	if err != nil {
		slog.Error("error: " + err.Error())
		return err
	}
	//Swap instruction
	swapIx, err = raydiumsdk.NewSwapV4Instruction(
		&r.swapInstruction,
		uint64(tokenAmount*math.Pow10(r.decimals)),
		0,
		solana.MustPublicKeyFromBase58(mintTokenAccount),
		wsolAccount,
		r.Wallet.PublicKey(),
	)
	if err != nil {
		slog.Error("error: " + err.Error())
		return err
	}

	// Create the transaction & Send.
	var instructions = []solana.Instruction{createATAWSOLSEED, initSeedWSOL, swapIx, closeWSOLATASEED}
	computeUnitLimitIx, _ := utils.ComputeBudgetLimit(130_000)
	computeBudgetPriceIx, _ := utils.ComputeBudgetPrice(uint64(r.FixedFee * 1e9))
	instructions = append([]solana.Instruction{computeBudgetPriceIx.Build(), computeUnitLimitIx.Build()}, instructions...)
	tipIx, err := r.Sender.GetTipInstructions(r.Wallet, config.CopyTraderConfig.SellTipFee)
	if err != nil {
		slog.Error("Error getting tip instructions: ", err)
		return err
	}
	instructions = append(instructions, tipIx...)
	tx, err := solana.NewTransaction(instructions, *r.Blockhash, solana.TransactionPayer(r.Wallet.PublicKey()))
	if err != nil {
		slog.Error("Error creating transaction", err)
	}
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &r.Wallet.PrivateKey
	})
	if err != nil {
		slog.Error("Error signing transaction ", err)
	}
	ok, err := r.Sender.SendTransaction(*tx, strategy.SendParams{
		TaskName:         "Raydium RiskManager",
		SimulationFailed: false,
	})
	if err != nil {
		slog.Error("Error sending transaction ", err)
	}
	if ok {
		slog.Info("Successfully sent transaction: " + tx.Signatures[0].String())
	} else {
		slog.Warn("Could not confirm transaction")
	}
	// --- snip ---
	r.active = false
	return nil
}

func (r *RaydiumTrade) GetTakeProfit() float64 {
	return r.TakeProfit
}
func (r *RaydiumTrade) GetPrice(opts *RPCOpts) (float64, error) {
	pricePerToken, _, _, _, _, err := raydiumsdk.GetPricePerToken(opts.Client, r.swapInstruction.GetAmmAccount().PublicKey.String(), 10, 50)
	if err != nil {
		return -1, err
	}
	return float64(pricePerToken), nil
}
func (r *RaydiumTrade) Active() bool {
	return r.active
}
func (r *RaydiumTrade) CalculateProfitLoss(currentPrice float64) float64 {
	if r.EntryPrice == 0 {
		return 10 // 10 is a placeholder value
	}
	// Calculate the profit or loss based on the current price.
	// if difference is 50% -> 1.5
	// if -50% -> 0.5
	return currentPrice / r.EntryPrice
}
func (r *RaydiumTrade) GetStopLoss() float64 {
	return r.StopLoss
}
func (r *RaydiumTrade) GetMint() string {
	return r.Mint
}
func (r *RaydiumTrade) GetExecutionOpts() ExecutionOpts {
	return r.ExecutionOpts
}
func (r *RaydiumTrade) SetActive(active bool) {
	r.active = active
}
