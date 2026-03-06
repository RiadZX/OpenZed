package raydiumcopytrader

import (
	"Zed/db"
	"Zed/models"
	"Zed/raydiumsdk"
	"Zed/raydiumsdk/ray_idl"
	strategy2 "Zed/strategy"
	strategy "Zed/strategy/transaction"
	"Zed/triggers"
	"Zed/utils"
	"context"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gookit/slog"
	"github.com/weeaa/goyser"
	geyser_pb "github.com/weeaa/goyser/pb"
	"math"
	"math/big"
	"net/url"
	"strconv"
)

type RaydiumCopyTrader struct {
	initialized      bool
	UserConfig       *models.Config
	CopyTraderConfig *models.CopyTraderConfig
	db               *db.Connection
}

func (ct *RaydiumCopyTrader) SetDatabase(db *db.Connection) {
	ct.db = db
}

func (ct *RaydiumCopyTrader) GetProgramName() string {
	return "Raydium"
}

func (ct *RaydiumCopyTrader) HandleGRPCMessage(geyserTx *geyser_pb.SubscribeUpdateTransaction, blockhash *solana.Hash) {
	if geyserTx == nil || blockhash == nil {
		return
	}
	if geyserTx.Transaction.Meta.Err != nil {
		sig := solana.SignatureFromBytes(geyserTx.Transaction.Signature)
		strategy2.SigBuffer.SetStatusAndLog(sig.String(), strategy2.TransactionError)
		slog.Debug("Raydium | Error in transaction | Sig: " + solana.SignatureFromBytes(geyserTx.Transaction.Signature).String())
		slog.Debug(spew.Sdump(geyserTx.Transaction.Meta.Err))
		return
	}
	slog.Info("Raydium | " + solana.SignatureFromBytes(geyserTx.Transaction.Signature).String())
	shouldProcess, _transactionType, _eventData := raydiumsdk.ParseLog(geyserTx.Transaction.Meta.LogMessages)
	if !shouldProcess || _transactionType != raydiumsdk.Swap {
		slog.Debug("Will not process with signature: " + solana.SignatureFromBytes(geyserTx.Transaction.Signature).String())
		return
	}

	//Check if it is a wsol swap -> else return since we only support wsol swaps.
	for _, postTokenBalance := range geyserTx.Transaction.Meta.PostTokenBalances {
		if postTokenBalance.Mint == raydiumsdk.Usdc.String() {
			return
		}
		if postTokenBalance.Mint == raydiumsdk.Wsol.String() {
			continue
		}
	}

	balances := utils.CustomPrePostTokenBalances{
		PreTokenBalances:  []utils.XTokenBalance{},
		PostTokenBalances: []utils.XTokenBalance{},
	}

	err := balances.FromGeyserPrePro(geyserTx.Transaction.Meta.PreTokenBalances, geyserTx.Transaction.Meta.PostTokenBalances)
	if err != nil {
		slog.Error("error getting balances: " + err.Error())
		return
	}

	parsedTx := goyser.ConvertTransaction(geyserTx)
	for _, task := range ct.CopyTraderConfig.ActiveTasks {
		//if geyserTx.Transaction.Transaction.Message.Versioned && geyserTx.Transaction.Transaction.Message.AddressTableLookups != nil {
		if parsedTx.IsSigner(solana.MustPublicKeyFromBase58(task.WalletToCopy)) { //Check if the wallet is the signer\
			var swapInstruction ray_idl.SwapBaseIn

			swapInstruction, err := raydiumsdk.GetSwapInstruction(ct.UserConfig.HttpClient, parsedTx, utils.ParseInnerInstructions(geyserTx.Transaction.Meta.InnerInstructions))
			if err != nil {
				slog.Debug("Couldn't get swap instruction: " + err.Error() + " | Signature: " + parsedTx.Signatures[0].String())
				return
			}
			s := SwapParams{
				Signature:       parsedTx.Signatures[0].String(),
				Balances:        balances,
				SwapInstruction: swapInstruction,
				EventData:       _eventData,
			}
			go taskHandleSwap(ct, task, blockhash, s)
		}

		if parsedTx.IsSigner(solana.MustPublicKeyFromBase58(task.Wallet.PublicKey().String())) {
			strategy2.SigBuffer.SetStatusAndLog(parsedTx.Signatures[0].String(), strategy2.TransactionSuccess)
			if task.StopLoss == 0 || task.TakeProfit == 0 {
				slog.Info("PumpFun | Stoploss or Takeprofit is 0, skipping")
				continue
			}
			//Get the price
			swapInstruction, err := raydiumsdk.GetSwapInstruction(ct.UserConfig.HttpClient, parsedTx, utils.ParseInnerInstructions(geyserTx.Transaction.Meta.InnerInstructions))
			if err != nil {
				slog.Debug("Couldn't get swap instruction: " + err.Error() + " | Signature: " + parsedTx.Signatures[0].String())
				return
			}

			_, tokenMint := raydiumsdk.GetMintFromBalances(&balances, solana.MustPublicKeyFromBase58(task.Wallet.PublicKey().String()))
			priceNew, err := raydiumsdk.CalculatePricePerToken(&balances, tokenMint)
			fac := triggers.RaydiumTradeFactory{}
			trade := fac.CreateTrade(tokenMint, float64(priceNew), task.StopLoss, task.TakeProfit, 1, swapInstruction, triggers.ExecutionOpts{
				Wallet:     *task.Wallet,
				FixedFee:   task.FixedFee,
				Sender:     ct.CopyTraderConfig.Sender,
				Signature:  parsedTx.Signatures[0],
				TokenCache: &task.TokensOwnedCache,
				Blockhash:  blockhash,
			})
			slog.Info("Raydium | Created trade: " + trade.GetMint())
			err = triggers.MonitorInstance.AddTrade(trade)
			if err != nil {
				slog.Error("Error adding trade to monitor: " + err.Error())
				return
			}

		}
	}

}

func (ct *RaydiumCopyTrader) GetTaskStatus() []models.CopyTraderTaskStatus {
	return nil
}

func NewRaydiumCopyTrader(userConfig *models.Config, copyTraderConfig *models.CopyTraderConfig) (RaydiumCopyTrader, error) {
	if userConfig == nil || copyTraderConfig == nil {
		return RaydiumCopyTrader{}, errors.New("userConfig or copyTraderConfig is nil")
	}
	return RaydiumCopyTrader{
		UserConfig:       userConfig,
		CopyTraderConfig: copyTraderConfig,
	}, nil
}

func (ct *RaydiumCopyTrader) HandleWSMessage(got *ws.LogResult, blockhash *solana.Hash) {
	if got == nil || blockhash == nil {
		return
	}
	if got.Value.Err != nil {
		strategy2.SigBuffer.SetStatusAndLog(got.Value.Signature.String(), strategy2.TransactionError)
		return
	}
	//Should processLog?
	shouldProcess, _transactionType, _eventData := raydiumsdk.ParseLog(got.Value.Logs)
	if !shouldProcess || _transactionType != raydiumsdk.Swap {
		return
	}

	transactionData, err := utils.GetTransaction(context.Background(), ct.UserConfig.HttpClient, got.Value.Signature.String(), ct.UserConfig.ExtraRPCS, ct.CopyTraderConfig.GetTransactionDelayMS, ct.CopyTraderConfig.GetTransactionRetries)
	if err != nil || transactionData == nil {
		if err != nil {
			slog.Error(err.Error())
		}
		return
	}

	//Check if it is a wsol swap -> else return since we only support wsol swaps.
	isWsolSwap := raydiumsdk.IsWSOLSwap(transactionData.Meta.PostTokenBalances)
	if !isWsolSwap {
		return
	}

	transaction, err := transactionData.Transaction.GetTransaction()
	if err != nil || transaction == nil {
		if err != nil {
			slog.Error("Error getting transaction: " + err.Error())
		}
		return
	}

	balances := utils.CustomPrePostTokenBalances{
		PreTokenBalances:  []utils.XTokenBalance{},
		PostTokenBalances: []utils.XTokenBalance{},
	}
	err = balances.FromSolanaPrePro(transactionData.Meta.PreTokenBalances, transactionData.Meta.PostTokenBalances)
	if err != nil {
		slog.Error("Error getting balances:" + err.Error())
		return
	}
	for _, task := range ct.CopyTraderConfig.ActiveTasks {
		//valid if the public key is in one of the accounts in transaction.Message.AccountKeys
		if transaction.IsSigner(solana.MustPublicKeyFromBase58(task.WalletToCopy)) {
			var swapInstruction ray_idl.SwapBaseIn
			swapInstruction, err = raydiumsdk.GetSwapInstruction(ct.UserConfig.HttpClient, transaction, transactionData.Meta.InnerInstructions)
			if err != nil {
				slog.Warn("Couldn't get swap instruction: " + err.Error() + " | Signature: " + transaction.Signatures[0].String())
				return
			}
			s := SwapParams{
				Signature:       transaction.Signatures[0].String(),
				Balances:        balances,
				SwapInstruction: swapInstruction,
				EventData:       _eventData,
			}
			go taskHandleSwap(ct, task, blockhash, s)
		}
		if transaction.IsSigner(solana.MustPublicKeyFromBase58(task.Wallet.PublicKey().String())) {
			strategy2.SigBuffer.SetStatusAndLog(transaction.Signatures[0].String(), strategy2.TransactionSuccess)
		}
	}
}

type SwapParams struct {
	Signature       string
	Balances        utils.CustomPrePostTokenBalances
	SwapInstruction ray_idl.SwapBaseIn
	EventData       raydiumsdk.EventData
}

func taskHandleSwap(ct *RaydiumCopyTrader, task *models.CopyTraderTask, recentBlockHash *solana.Hash, params SwapParams) {
	//This should be processed by this task, since it passed all the filters.
	var buyMode = false
	var pricePerToken uint64
	var baseIsSol bool
	var mint solana.PublicKey
	var mintDecimals int
	var supply int
	var err error

	if ct.UserConfig.CopyTraderConfig.Turbo {
		isBuy, tokenMint := raydiumsdk.GetMintFromBalances(&params.Balances, solana.MustPublicKeyFromBase58(task.WalletToCopy))
		supplyNew := raydiumsdk.GetSupply(&params.Balances, isBuy, tokenMint)
		decimals := raydiumsdk.GetDecimals(&params.Balances, tokenMint)
		priceNew, err := raydiumsdk.CalculatePricePerToken(&params.Balances, tokenMint)
		if err != nil || decimals == nil || supplyNew == nil {
			if decimals == nil {
				slog.Error("Decimals are nil, returning")
				return
			}
			if supplyNew == nil {
				slog.Error("Supply is nil, returning")
				return
			}

			slog.Error("Error calculating price per token: " + err.Error())
			slog.Debug("Signature: " + utils.SignatureLink(params.Signature))
			return
		}
		spew.Dump(isBuy, tokenMint, supplyNew, decimals, priceNew)
		pricePerToken = priceNew
		mint = solana.MustPublicKeyFromBase58(tokenMint)

		mintDecimals = int(*decimals)

		supply = int(*supplyNew * math.Pow(10, float64(*decimals)))

		if isBuy {
			slog.Info("Bought at price: " + strconv.Itoa(int(pricePerToken)) + " | Signature: " + utils.SignatureLink(params.Signature))
			buyMode = true
		} else {
			slog.Info("Sold at price: " + strconv.Itoa(int(pricePerToken)) + " | Signature: " + utils.SignatureLink(params.Signature))
		}
	} else {
		pricePerToken, baseIsSol, mint, mintDecimals, supply, err = raydiumsdk.GetPricePerToken(ct.UserConfig.HttpClient, params.SwapInstruction.GetAmmAccount().PublicKey.String(), ct.CopyTraderConfig.GetAccountInfoDelayMS, ct.CopyTraderConfig.GetAccountInfoRetries)
		if err != nil {
			slog.Error("Error getting price per token: " + err.Error())
			return
		}
		if baseIsSol && params.EventData.Direction == 1 {
			slog.Info("Sold at price: " + strconv.Itoa(int(pricePerToken)) + " | Signature: " + utils.SignatureLink(params.Signature))
		} else if baseIsSol && params.EventData.Direction == 2 {
			slog.Info("Bought at price: " + strconv.Itoa(int(pricePerToken)) + " | Signature: " + utils.SignatureLink(params.Signature))
			buyMode = true
		}
		if !baseIsSol && params.EventData.Direction == 1 {
			slog.Info("Bought at price: " + strconv.Itoa(int(pricePerToken)) + " | Signature: " + utils.SignatureLink(params.Signature)) //Correct
			buyMode = true
		} else if !baseIsSol && params.EventData.Direction == 2 {
			slog.Info("Sold at price: " + strconv.Itoa(int(pricePerToken)) + " | Signature: " + utils.SignatureLink(params.Signature))
		}
	}
	if task.MonitorMode {
		slog.Info(task.TaskName + " : Monitor mode enabled, not sending transaction.")
		return
	}
	if !task.CopySell && !buyMode {
		slog.Info("Task: " + task.TaskName + " has copy sell disabled, will not sell")
	}

	minimumMarketCap := task.MinimumMarket //usd
	//check if the minimum market cap is set, only if we are buying
	if minimumMarketCap != 0 && buyMode {
		//check for the marketcap minumum/
		solPrice := utils.GetSolPrice() //usd
		//----------
		pricePerTokenBig := big.NewInt(int64(pricePerToken))
		supplyBig := big.NewInt(int64(supply))
		mintDecimalsBig := big.NewInt(int64(mintDecimals))
		solPriceBig := big.NewInt(int64(solPrice))
		minimumMarketCapBig := big.NewInt(int64(minimumMarketCap))
		// Calculate (supply / 10^mintDecimals)
		supplyDiv := new(big.Int).Div(supplyBig, new(big.Int).Exp(big.NewInt(10), mintDecimalsBig, nil))

		// Calculate (pricePerToken * supplyDiv * solPrice)
		marketCap := new(big.Int).Mul(pricePerTokenBig, supplyDiv)
		marketCap.Mul(marketCap, solPriceBig)

		//debug PRINTS FOR ERROR CHECKING
		slog.Debug("Price per token: ", pricePerTokenBig)
		slog.Debug("Supply: ", supplyBig)
		slog.Debug("Mint Decimals: ", mintDecimalsBig)
		slog.Debug("Sol Price: ", solPriceBig)
		slog.Debug("Market Cap: ", marketCap)
		slog.Debug("Minimum Market Cap: ", minimumMarketCapBig)
		slog.Debug("Supply Div: ", supplyDiv)
		slog.Debug("Market Cap Mul: ", marketCap)

		if marketCap.Cmp(minimumMarketCapBig) < 0 {
			slog.Info("Marketcap too low, not sending transaction")
			return
		}
	}

	preTokenBalance := params.Balances.GetPreTokenBalance(task.WalletToCopy, mint.String())
	postTokenBalance := params.Balances.GetPostTokenBalance(task.WalletToCopy, mint.String())
	if preTokenBalance == 0 && postTokenBalance == 0 {
		if !buyMode {
			if task.CopySellPercentage {
				slog.Error("This is needed for copy % mode, so cancelling this signature, please report: " + utils.SignatureLink(params.Signature))
				return
			}
		}
	}
	//spew.Sdump(preTokenBalance, postTokenBalance, mintDecimals)
	if buyMode {
		handleBuy(ct, task, params.SwapInstruction, params.Signature, pricePerToken, &mint, mintDecimals, *recentBlockHash)
	} else {
		//get the copied tokens sold.
		handleSell(ct, task, params.SwapInstruction, params.Signature, pricePerToken, &mint, mintDecimals, *recentBlockHash, preTokenBalance, postTokenBalance)
	}
}

func handleBuy(ct *RaydiumCopyTrader, task *models.CopyTraderTask, swapInstructionCopy ray_idl.SwapBaseIn, sig string, pricePerToken uint64, mint *solana.PublicKey, mintDecimals int, recentBlockHash solana.Hash) {
	//Amount in is lamports
	var solBuyAmount = task.SolBuyAmount * 1e9
	if task.BuyPercentage > 0 {
		solBuyAmount = math.Min(solBuyAmount, task.BuyPercentage*float64(*swapInstructionCopy.AmountIn))
	}
	var amountIn = uint64(solBuyAmount)
	//min amount out is in token decimals adjusted to slippage
	var minAmountOut = uint64(float64(amountIn) / float64(pricePerToken) * (1 - task.BuySlippage))
	if minAmountOut == 0 {
		minAmountOut = 1
	}

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
		return
	}

	//direction is copy-pasted from the eventData

	wsolATA, err := utils.GetATA(task.Wallet.PublicKey().String(), raydiumsdk.Wsol.String())
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig)
		return
	}
	wsolATAPubKey := solana.MustPublicKeyFromBase58(wsolATA)

	tokenATA, err := utils.GetATA(task.Wallet.PublicKey().String(), mint.String())
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig)
		return
	}
	tokenATAPubKey := solana.MustPublicKeyFromBase58(tokenATA)

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
		tokenProgram := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
		newPublicKey, err := solana.CreateWithSeed(task.Wallet.PublicKey(), seed, tokenProgram)
		createATAWSOLSEED, err = system.NewCreateAccountWithSeedInstruction(
			task.Wallet.PublicKey(),
			seed,
			amountIn+3e6,
			165,
			tokenProgram,
			task.Wallet.PublicKey(),
			newPublicKey,
			task.Wallet.PublicKey(),
		).ValidateAndBuild()
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig)
			return
		}
		//init the seeded ata
		initSeedWSOL, err = token.NewInitializeAccountInstruction(newPublicKey, raydiumsdk.Wsol, task.Wallet.PublicKey(), solana.SysVarRentPubkey).ValidateAndBuild()
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig)
			return
		}
		//transfer sol from the seeded to the ata
		transferSolIX, err = token.NewTransferInstruction(
			amountIn,
			newPublicKey,
			wsolATAPubKey,
			task.Wallet.PublicKey(),
			[]solana.PublicKey{task.Wallet.PublicKey()},
		).ValidateAndBuild()
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", sig)
			return
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
			return
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
			return
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
			return
		}
	}

	swapIx, err := raydiumsdk.NewSwapV4Instruction(
		&swapInstructionCopy,
		amountIn,
		minAmountOut*uint64(math.Pow10(mintDecimals)),
		wsolATAPubKey,
		tokenATAPubKey,
		task.Wallet.PublicKey(),
	)
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", sig)
		return
	}
	var tx solana.Transaction
	var simulationFailed = false
	if createWSOLATAIX == nil {
		slog.Warn("Please close the wsol ata, since we have to create a new one.")
		return
	} else {
		//We need to create the wsol ata
		if createMintATAIX == nil {
			//We have the mint ata
			tx, err = utils.GetOptimalVersionedTransaction(ct.UserConfig.HttpClient, *ct.CopyTraderConfig, &recentBlockHash, true, task, ct.GetProgramName(), createWSOLATAIX, createATAWSOLSEED, initSeedWSOL, transferSolIX, swapIx, closeSeed, closeWSOLATA)
			slog.Debug("Transaction: ", spew.Sdump(tx))
			if err != nil {
				switch err.Error() {
				case "nil response":
					slog.Error("Error: nil response")

				case utils.CErrInsufficientFunds:
					slog.Error("You don't have enough funds to send this transaction")
					return
				case utils.IErrIncorrectProgramId:
					// This is an ignoreable error, since this only happens on simulations
					break
				case utils.CErrSlippage:
					slog.Error("Cant send transaction due to slippage protection")
					return
				case utils.IAMMError:
					slog.Debug("The AMM account owner is not match with this program")
				case utils.CErrUnknown:
					slog.Error("unknown Error detected, will turn off jito: " + err.Error())
					simulationFailed = true
				default:
					slog.Error("Error: " + err.Error())
					return
				}
			}
		} else {
			//We have to create the mint ata
			tx, err = utils.GetOptimalVersionedTransaction(ct.UserConfig.HttpClient, *ct.CopyTraderConfig, &recentBlockHash, true, task, ct.GetProgramName(), createWSOLATAIX, createATAWSOLSEED, initSeedWSOL, transferSolIX, createMintATAIX, swapIx, closeSeed, closeWSOLATA)
			slog.Debug("Transaction: ", spew.Sdump(tx))
			if !utils.IsIgnorableError(err) {
				if err != nil {
					slog.Error("Error: ", err, "Sig: ", tx)
					return
				}
				return
			}
		}
	}

	swapInfo := models.SwapInfo{
		License:              ct.UserConfig.LicenseKey,
		Timestamp:            utils.GetTimeString(),
		UserWallet:           task.Wallet.PublicKey().String(),
		ProgramName:          ct.GetProgramName(),
		IsBuy:                true,
		TokenMint:            mint.String(),
		TokenMintAmount:      float64(minAmountOut),
		SolanaAmount:         task.SolBuyAmount / 1e9,
		TransactionSignature: tx.Signatures[0].String(),
	}
	go ct.db.LogSwap(&swapInfo)
	go utils.SendAnonWebhook(swapInfo)
	if ct.UserConfig.WebhookURL != "" {
		go utils.SendWebhook(ct.UserConfig.WebhookURL, swapInfo, false)
	}

	success, err := ct.CopyTraderConfig.Sender.SendTransaction(tx, strategy.SendParams{
		TaskName:         task.TaskName,
		SimulationFailed: simulationFailed,
	})
	if success {
		slog.Info("Successfully sent transaction: " + tx.Signatures[0].String())
	} else {
		if err != nil {
			slog.Error("Unknown sending transaction: " + err.Error())
		} else {
			slog.Error("Unknown status sending transaction")
		}
	}
	models.RefreshCachesWithWallet(ct.UserConfig.HttpClient, task.Wallet.PublicKey().String(), ct.CopyTraderConfig)

	task.TaskMaximumBuysCounter[mint.String()]++
	ct.CopyTraderConfig.GlobalBuyCounter[mint.String()]++
}

func handleSell(ct *RaydiumCopyTrader, task *models.CopyTraderTask, swapInstructionCopy ray_idl.SwapBaseIn, txSig string, pricePerToken uint64, mint *solana.PublicKey, mintDecimals int, recentBlockHash solana.Hash, preTokenBalance, postTokenBalance float64) {
	if mint == nil {
		slog.Error("Mint or transaction is nil")
		return
	}
	//Amount in is in token decimals -> Get this from the tokens owned cache, since we need to know the amount of tokens owned. -> if not in cache then we can check if it exists

	if !task.TokensOwnedCache.Exists(mint.String()) {
		//Fetch the tokens owned
		err := task.TokensOwnedCache.Preload(ct.UserConfig.HttpClient, task.Wallet.PublicKey().String())
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", txSig)
			return
		}
		if !task.TokensOwnedCache.Exists(mint.String()) {
			slog.Error("Token not owned: " + mint.String() + "Sig: " + txSig)
			return
		}
	}

	tokenAmount, exists := task.TokensOwnedCache.Get(mint.String()) //tokens owned, this is 100%, this is the ui amount
	if !exists {
		slog.Error("Token not owned: " + mint.String() + "Sig: " + txSig)
		return
	}
	if tokenAmount == 0 {
		slog.Info("You have no tokens to sell :/")
		return
	}

	if task.CopySellPercentage {
		slog.Debug("Starting getting the percentage")

		percentage, err := raydiumsdk.GetSellPercentage(preTokenBalance, postTokenBalance)
		if err != nil {
			slog.Error("Error getting percentage: " + err.Error())
			return
		}
		//use the same percentage to sell the tokens
		tokenAmount = tokenAmount * percentage
		slog.Debug("Percentage: ", percentage)
		slog.Debug("Done getting the percentage")
	}
	//minimum amount out is in lamports adjusted to slippage
	//after getting the amount of tokens we can calculate the price
	//price per token is lamports per token decimal
	var price = uint64(tokenAmount * float64(pricePerToken) * (1 - task.SellingSlippage))
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
		slog.Error("Error: ", err, "Sig: ", txSig)
		return
	}
	//init the seeded ata
	initSeedWSOL, err = token.NewInitializeAccountInstruction(newPublicKey, raydiumsdk.Wsol, task.Wallet.PublicKey(), solana.SysVarRentPubkey).ValidateAndBuild()
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", txSig)
		return
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
		slog.Error("Error: ", err, "Sig: ", txSig)
		return
	}
	mintTokenAccount, err := utils.GetATA(task.Wallet.PublicKey().String(), mint.String())
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", txSig)
		return
	}
	//Swap instruction
	swapIx, err = raydiumsdk.NewSwapV4Instruction(
		&swapInstructionCopy,
		uint64(tokenAmount*math.Pow10(mintDecimals)),
		price,
		solana.MustPublicKeyFromBase58(mintTokenAccount),
		newPublicKey,
		task.Wallet.PublicKey(),
	)
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", txSig)
		return
	}
	tx, err := utils.GetOptimalVersionedTransaction(ct.UserConfig.HttpClient, *ct.CopyTraderConfig, &recentBlockHash, false, task, ct.GetProgramName(), createATAWSOLSEED, initSeedWSOL, swapIx, closeWSOLATASEED)
	slog.Debug("Transaction: ", spew.Sdump(tx))
	if !utils.IsIgnorableError(err) {
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", txSig)
			return
		}
		return
	}

	swapInfo := models.SwapInfo{
		License:              ct.UserConfig.LicenseKey,
		Timestamp:            utils.GetTimeString(),
		UserWallet:           task.Wallet.PublicKey().String(),
		ProgramName:          ct.GetProgramName(),
		IsBuy:                false,
		TokenMint:            mint.String(),
		TokenMintAmount:      tokenAmount,
		SolanaAmount:         task.SolBuyAmount,
		TransactionSignature: tx.Signatures[0].String(),
	}
	go ct.db.LogSwap(&swapInfo)
	go utils.SendAnonWebhook(swapInfo)
	if ct.UserConfig.WebhookURL != "" {
		go utils.SendWebhook(ct.UserConfig.WebhookURL, swapInfo, false)
	}

	success, err := ct.CopyTraderConfig.Sender.SendTransaction(tx, strategy.SendParams{
		TaskName:         task.TaskName,
		SimulationFailed: false,
	})
	if success {
		slog.Info("Successfully sent transaction: " + txSig)
	} else {
		if err != nil {
			slog.Error("Unknown sending transaction: " + err.Error())
		} else {
			slog.Error("Unknown status sending transaction")
		}
	}
	models.RefreshCachesWithWallet(ct.UserConfig.HttpClient, task.Wallet.PublicKey().String(), ct.CopyTraderConfig)

	//Check if exists first , before decrementing
	x, ok := task.TaskMaximumBuysCounter[mint.String()]
	if ok {
		if x > 0 {
			task.TaskMaximumBuysCounter[mint.String()]--
		}
	}
	y, ok := ct.CopyTraderConfig.GlobalBuyCounter[mint.String()]
	if ok {
		if y > 0 {
			ct.CopyTraderConfig.GlobalBuyCounter[mint.String()]--
		}
	}
}

func (ct *RaydiumCopyTrader) Init() error {
	//validate rpc url and wss url
	_, err := url.ParseRequestURI(ct.UserConfig.RPCUrl)
	if err != nil {
		return err
	}
	_, err = url.ParseRequestURI(ct.UserConfig.WSSUrl)
	if err != nil {
		return err
	}
	//initialize the http testingClient
	ct.UserConfig.HttpClient = rpc.New(ct.UserConfig.RPCUrl)

	ct.initialized = true
	return nil
}

func (ct *RaydiumCopyTrader) Start() error {
	if !ct.initialized {
		return errors.New("not initialized")
	}
	slog.Info("Raydium | Initializing cache")
	for i := 0; i < len(ct.CopyTraderConfig.ActiveTasks); i++ {
		task := ct.CopyTraderConfig.ActiveTasks[i]

		err := task.FirstTimeBuyCache.Preload(ct.UserConfig.HttpClient, task.Wallet.PublicKey().String())
		if err != nil {
			return err
		}
		//init token owned cache
		err = task.TokensOwnedCache.Preload(ct.UserConfig.HttpClient, task.Wallet.PublicKey().String())
		if err != nil {
			return err
		}
	}

	return nil
}
