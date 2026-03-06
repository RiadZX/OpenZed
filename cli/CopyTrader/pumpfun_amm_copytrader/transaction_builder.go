package pumpfun_amm_copytrader

import (
	"Zed/models"
	"Zed/pumpfunsdk"
	"Zed/raydiumsdk"
	"Zed/utils"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gookit/slog"
	"math"
	"math/big"
)

// BuildBuyTransaction builds a buy transaction.
// This builds the full transaction, including fee related instructions, handling wsol/ata accounts, etc.
// The returned transaction can be used to send directly to the network.
// This will also use the caches, and everything needed to build the FULL transaction.
// This is similar to the Raydium tx flow.
func BuildBuyTransaction(
	ct *PfAmmCt,
	task *models.CopyTraderTask,
	recentBlockHash solana.Hash,
	sig solana.Signature,
	ammEvent *pumpfunsdk.AMMBuyEvent,
	ammSwapTransaction *pumpfunsdk.AMMSwapTransaction,
	user_base_token_account solana.PublicKey,
	user_quote_token_account solana.PublicKey,
	protocol_fee_recipient solana.PublicKey,
	protocol_fee_recipient_token_account solana.PublicKey,
) (*solana.Transaction, error) {
	// Amount in is lamports
	var QuoteAmountIn = uint64(task.SolBuyAmount * 1e9)
	if task.BuyPercentage > 0 {
		QuoteAmountIn = uint64(math.Min(float64(QuoteAmountIn), task.BuyPercentage*float64(ammEvent.Quote_amount_in)))
	}
	maxQuoteAmountIn := uint64(float64(QuoteAmountIn) * (1 + task.BuySlippage))

	// Calculate total fee basis points and denominator
	lpFeeBps := big.NewInt(int64(ammEvent.Lp_fee_basis_points))
	protocolFeeBps := big.NewInt(int64(ammEvent.Protocol_fee_basis_points))
	totalFeeBps := new(big.Int).Add(lpFeeBps, protocolFeeBps)
	denominator := new(big.Int).Add(big.NewInt(10000), totalFeeBps)

	// Calculate effective quote amount
	effectiveQuote := new(big.Int).Mul(big.NewInt(int64(QuoteAmountIn)), big.NewInt(10000))
	effectiveQuote = effectiveQuote.Div(effectiveQuote, denominator)

	// Calculate the base tokens received using effectiveQuote
	numerator := new(big.Int).Mul(big.NewInt(int64(ammEvent.Pool_base_token_reserves)), effectiveQuote)
	denominatorEffective := new(big.Int).Add(big.NewInt(int64(ammEvent.Pool_quote_token_reserves)), effectiveQuote)

	if denominatorEffective.Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New("Pool would be depleted; denominator is zero.")
	}

	baseAmountOut := new(big.Int).Div(numerator, denominatorEffective).Uint64()

	//slog.Info("Base amount out: ", baseAmountOut)
	//slog.Info("Original base amount out: ", ammEvent.Base_amount_out)

	mint := ammSwapTransaction.BuyTransaction.BaseMint
	//Check if allowed to buy according to the token counter
	taskAmount, ok := task.TaskMaximumBuysCounter[mint.String()]
	globalAmount, ok2 := ct.CopyTraderConfig.GlobalBuyCounter[mint.String()]
	if !ok {
		taskAmount = 0
		task.TaskMaximumBuysCounter[mint.String()] = 0
	}
	if !ok2 {
		globalAmount = 0
		ct.CopyTraderConfig.GlobalBuyCounter[mint.String()] = 0
	}
	if taskAmount >= task.TaskMaximumBuys || globalAmount >= ct.CopyTraderConfig.MaximumBuysPerToken || taskAmount >= ct.CopyTraderConfig.MaximumBuysPerToken {
		slog.Info("Task: " + task.TaskName + " has reached the maximum buys for this token")
		return nil, errors.New("Task: " + task.TaskName + " has reached the maximum buys for this token")
	}

	//direction is copy-pasted from the eventData
	wsolATA, err := utils.GetATA(task.Wallet.PublicKey().String(), raydiumsdk.Wsol.String())
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig)
		return nil, err
	}
	wsolATAPubKey := solana.MustPublicKeyFromBase58(wsolATA)
	tokenATA, err := utils.GetATA(task.Wallet.PublicKey().String(), mint.String())
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig)
		return nil, err
	}

	var createWSOLATAIX *associatedtokenaccount.Instruction
	var createATAWSOLSEED *system.Instruction
	var initSeedWSOL *token.Instruction
	var transferSolIX *token.Instruction
	var closeSeed *token.Instruction
	var closeWSOLATA *token.Instruction

	if !task.FirstTimeBuyCache.Exists(wsolATA) {
		//Create the ata
		createWSOLATAIX = associatedtokenaccount.NewCreateInstruction(task.Wallet.PublicKey(), task.Wallet.PublicKey(), raydiumsdk.Wsol).Build()
		//creating the ata seed account
		seed := solana.NewWallet().PublicKey().String()[0:32]
		//concat the public key + seed + program id into sha256 to create a new public key
		newPublicKey, err := solana.CreateWithSeed(task.Wallet.PublicKey(), seed, solana.TokenProgramID)
		createATAWSOLSEED, err = system.NewCreateAccountWithSeedInstruction(
			task.Wallet.PublicKey(),
			seed,
			maxQuoteAmountIn+3e6,
			165,
			solana.TokenProgramID,
			task.Wallet.PublicKey(),
			newPublicKey,
			task.Wallet.PublicKey(),
		).ValidateAndBuild()
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig)
			return nil, err
		}
		//init the seeded ata
		initSeedWSOL, err = token.NewInitializeAccountInstruction(newPublicKey, raydiumsdk.Wsol, task.Wallet.PublicKey(), solana.SysVarRentPubkey).ValidateAndBuild()
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig)
			return nil, err
		}
		//transfer sol from the seeded to the ata
		transferSolIX, err = token.NewTransferInstruction(
			maxQuoteAmountIn,
			newPublicKey,
			wsolATAPubKey,
			task.Wallet.PublicKey(),
			[]solana.PublicKey{task.Wallet.PublicKey()},
		).ValidateAndBuild()
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig)
			return nil, err
		}

		//close the seed account
		closeSeed, err = token.NewCloseAccountInstruction(
			newPublicKey,
			task.Wallet.PublicKey(),
			task.Wallet.PublicKey(),
			[]solana.PublicKey{task.Wallet.PublicKey()},
		).ValidateAndBuild()
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig)
			return nil, err
		}

		//close the wosl ata
		closeWSOLATA, err = token.NewCloseAccountInstruction(
			wsolATAPubKey,
			task.Wallet.PublicKey(),
			task.Wallet.PublicKey(),
			[]solana.PublicKey{task.Wallet.PublicKey()},
		).ValidateAndBuild()
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig)
			return nil, err
		}
	}

	//Token account for the mint
	var createMintATAIX solana.Instruction
	if !task.FirstTimeBuyCache.Exists(tokenATA) {
		//Create the ata idempotent
		fmt.Println("Creating mint ata")
		createMintATAIX, err = utils.ConstructCreateATAInstruction(task.Wallet.PublicKey().String(), mint.String())
		if err != nil {
			slog.Error("Error constructing the idempotent create ata instruction for createmintataix", err, "Sig: ", sig)
			return nil, err
		}
	}

	//Create the buy instruction on the amm
	buyIx, err := pumpfunsdk.NewAMMBuyInstruction(
		maxQuoteAmountIn,
		baseAmountOut,
		*ammEvent,
		*ammSwapTransaction.BuyTransaction,
		user_base_token_account,
		user_quote_token_account,
		protocol_fee_recipient,
		protocol_fee_recipient_token_account,
		task.Wallet.PublicKey(),
	)
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig)
		return nil, err
	}

	// print all instructions
	slog.Debug("📜 AMM Buy Transaction:")
	slog.Debug(spew.Sdump(buyIx))
	slog.Debug("createWSOLATAIX: ")
	slog.Debug(spew.Sdump(createWSOLATAIX))
	slog.Debug("createATAWSOLSEED: ")
	slog.Debug(spew.Sdump(createATAWSOLSEED))
	slog.Debug("initSeedWSOL: ")
	slog.Debug(spew.Sdump(initSeedWSOL))
	slog.Debug("transferSolIX: ")
	slog.Debug(spew.Sdump(transferSolIX))
	slog.Debug("closeSeed: ")
	slog.Debug(spew.Sdump(closeSeed))
	slog.Debug("closeWSOLATA: ")
	slog.Debug(spew.Sdump(closeWSOLATA))
	slog.Debug("createMintATAIX: ")
	slog.Debug(spew.Sdump(createMintATAIX))

	//Create the transaction
	//simulationFailed := false
	var tx solana.Transaction
	if createWSOLATAIX == nil {
		slog.Warn("Please close the wsol ata, since we have to create a new one.")
		return nil, errors.New("please close the wsol ata, since we have to create a new one")
	}
	if createMintATAIX != nil {
		tx, err = utils.GetOptimalVersionedTransaction(ct.UserConfig.HttpClient, *ct.CopyTraderConfig, &recentBlockHash, true, task, ct.GetProgramName(), createWSOLATAIX, createATAWSOLSEED, initSeedWSOL, transferSolIX, createMintATAIX, buyIx, closeSeed, closeWSOLATA)

	} else {
		tx, err = utils.GetOptimalVersionedTransaction(ct.UserConfig.HttpClient, *ct.CopyTraderConfig, &recentBlockHash, true, task, ct.GetProgramName(), createWSOLATAIX, createATAWSOLSEED, initSeedWSOL, transferSolIX, buyIx, closeSeed, closeWSOLATA)
	}
	slog.Debug("Transaction: ", spew.Sdump(tx))
	if err != nil {
		switch err.Error() {
		case "nil response":
			slog.Error("Error: nil response")

		case utils.CErrInsufficientFunds:
			slog.Error("You don't have enough funds to send this transaction")
			return nil, err
		case utils.IErrIncorrectProgramId:
			// This is an ignoreable error, since this only happens on simulations
			break
		case utils.CErrSlippage:
			slog.Error("Cant send transaction due to slippage protection")
			return nil, err
		case utils.IAMMError:
			slog.Debug("The AMM account owner is not match with this program")
		case utils.CErrUnknown:
			slog.Error("unknown Error detected, will turn off jito: " + err.Error())
			//simulationFailed = true
			break
		default:
			slog.Error("Error: " + err.Error())
			return nil, err
		}
	}

	return &tx, nil
}

// BuildSellTransaction builds a sell transaction.
// This builds the full transaction, including fee related instructions, handling wsol/ata accounts, etc.
// The returned transaction can be used to send directly to the network.
// This will also use the caches, and everything needed to build the FULL transaction.
// This is similar to the Raydium tx flow.
func BuildSellTransaction(
	ct *PfAmmCt,
	task *models.CopyTraderTask,
	recentBlockHash solana.Hash,
	sig solana.Signature,
	ammEvent *pumpfunsdk.AMMSellEvent,
	ammSwapTransaction *pumpfunsdk.AMMSwapTransaction,
	user_base_token_account solana.PublicKey,
	protocol_fee_recipient solana.PublicKey,
	protocol_fee_recipient_token_account solana.PublicKey,
	decimals int,
) (*solana.Transaction, error) {
	mint := ammSwapTransaction.SellTransaction.BaseMint
	//Amount in is in token decimals -> Get this from the tokens owned cache, since we need to know the amount of tokens owned. -> if not in cache then we can check if it exists

	if !task.TokensOwnedCache.Exists(mint.String()) {
		//Fetch the tokens owned
		err := task.TokensOwnedCache.Preload(ct.UserConfig.HttpClient, task.Wallet.PublicKey().String())
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig.String())
			return nil, err
		}
		if !task.TokensOwnedCache.Exists(mint.String()) {
			slog.Error("Token not owned: " + mint.String() + "Sig: " + sig.String())
			return nil, errors.New("Token not owned: " + mint.String() + "Sig: " + sig.String())
		}
	}

	baseAmountIn, exists := task.TokensOwnedCache.Get(mint.String()) //tokens owned, this is 100%, this is the ui amount
	if !exists {
		slog.Error("Token not owned: " + mint.String() + "Sig: " + sig.String())
		return nil, errors.New("Token not owned: " + mint.String() + "Sig: " + sig.String())
	}
	if baseAmountIn == 0 {
		slog.Warn("You have no tokens to sell :/")
		return nil, errors.New("You have no tokens to sell :/")
	}
	//baseAmountIn = 1_000_000_000 //TODO: remove this, this is just for testing

	if task.CopySellPercentage {
		slog.Warn("Copy sell percentage is currently disabled for amm swap. if you want to use it, please contact the devs.")
	}

	var minQuoteAmountOut = 0

	//Instructions needed
	//Create account with seed for wsol
	//initialize the seeded wsol
	//swap ix
	//close the seeded sol.
	var createATAWSOLSEED *system.Instruction
	var initSeedWSOL *token.Instruction
	var closeWSOLATASEED *token.Instruction
	//creating the ata seed account
	seed := solana.NewWallet().PublicKey().String()[0:32]
	//concat the public key + seed + program id into sha256 to create a new public key
	tokenProgram := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	newPublicKey, err := solana.CreateWithSeed(task.Wallet.PublicKey(), seed, tokenProgram)
	createATAWSOLSEED, err = system.NewCreateAccountWithSeedInstruction(
		task.Wallet.PublicKey(),
		seed,
		3e6,
		165,
		tokenProgram,
		task.Wallet.PublicKey(),
		newPublicKey,
		task.Wallet.PublicKey(),
	).ValidateAndBuild()
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig.String())
		return nil, err
	}
	//init the seeded ata
	initSeedWSOL, err = token.NewInitializeAccountInstruction(newPublicKey, raydiumsdk.Wsol, task.Wallet.PublicKey(), solana.SysVarRentPubkey).ValidateAndBuild()
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig.String())
		return nil, err
	}

	//close the wsol seed ata ix
	//close the seed account
	closeWSOLATASEED, err = token.NewCloseAccountInstruction(
		newPublicKey,
		task.Wallet.PublicKey(),
		task.Wallet.PublicKey(),
		[]solana.PublicKey{task.Wallet.PublicKey()},
	).ValidateAndBuild()
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig.String())
		return nil, err
	}

	//Swap instruction
	sellIx, err := pumpfunsdk.NewAMMSellInstruction(
		uint64(baseAmountIn*math.Pow(10, float64(decimals))),
		uint64(minQuoteAmountOut),
		*ammEvent,
		*ammSwapTransaction.SellTransaction,
		user_base_token_account,
		newPublicKey,
		protocol_fee_recipient,
		protocol_fee_recipient_token_account,
		task.Wallet.PublicKey(),
	)

	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig.String())
		return nil, err
	}
	tx, err := utils.GetOptimalVersionedTransaction(ct.UserConfig.HttpClient, *ct.CopyTraderConfig, &recentBlockHash, false, task, ct.GetProgramName(), createATAWSOLSEED, initSeedWSOL, sellIx, closeWSOLATASEED)
	slog.Debug("Transaction: ", spew.Sdump(tx))
	if !utils.IsIgnorableError(err) {
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig.String())
			return nil, err
		}
		return nil, errors.New("Error: nil response")
	}

	return &tx, nil
}
