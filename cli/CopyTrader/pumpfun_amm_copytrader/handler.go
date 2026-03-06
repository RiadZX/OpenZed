package pumpfun_amm_copytrader

import (
	"Zed/models"
	"Zed/pumpfunsdk"
	strategy "Zed/strategy/transaction"
	"Zed/utils"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"math/rand"
)

func Handle(ct *PfAmmCt, task *models.CopyTraderTask, recentBlockHash solana.Hash, txSig solana.Signature, ammEvent *pumpfunsdk.AMMEvent, tx *solana.Transaction, innerInstructions *[]rpc.InnerInstruction, balances utils.CustomPrePostTokenBalances) {
	// Print the details of the amm event in pretty format with emojis to the console
	//slog.Info("PumpFun AMM | Handling AMM event | Sig: " + txSig.String())
	slog.Info("📜 AMM Event Data:")
	if ammEvent.IsBuy {
		slog.Info("📈 Buy")
		slog.Info(fmt.Sprintf("🥽 Pool Address: %s", ammEvent.Buy.Pool.String()))
		slog.Info(fmt.Sprintf("💲 Base Amount Out: %d", ammEvent.Buy.Base_amount_out))
		slog.Info(fmt.Sprintf("💱 Max Quote Amount In: %d", ammEvent.Buy.Max_quote_amount_in))

	} else {
		slog.Info("📉 Sell")
		slog.Info(fmt.Sprintf("🥽 Pool Address: %s", ammEvent.Sell.Pool.String()))
		slog.Info(fmt.Sprintf("💲 Base Amount In: %d", ammEvent.Sell.Base_amount_in))
		slog.Info(fmt.Sprintf("💱 Min Quote Amount Out: %d", ammEvent.Sell.Min_quote_amount_out))
	}

	// We know that we should buy or sell the token in the amm event.
	tx, err := utils.ResolveAddressTables(ct.UserConfig.HttpClient, tx)
	if err != nil {
		slog.Error("Error resolving address tables: " + err.Error() + " Signature: " + txSig.String())
		return
	}
	// Detect if it is a buy or sell event
	if ammEvent.IsBuy {
		HandleBuy(ct, task, recentBlockHash, txSig, ammEvent, tx, innerInstructions, balances)
	} else {
		HandleSell(ct, task, recentBlockHash, txSig, ammEvent, tx, innerInstructions, balances)
	}
}

func HandleBuy(ct *PfAmmCt, task *models.CopyTraderTask, recentBlockHash solana.Hash, txSig solana.Signature, ammEvent *pumpfunsdk.AMMEvent, tx *solana.Transaction, innerInstructions *[]rpc.InnerInstruction, balances utils.CustomPrePostTokenBalances) {
	// We know that we should buy the token in the amm event.

	// this is a buy. so basemint = mint , quote mint = wsol
	ammSwapTransaction, err := pumpfunsdk.ParseAMMTransaction(tx, innerInstructions)
	if err != nil {
		slog.Error("Error parsing AMM transaction: " + err.Error())
		return
	}

	slog.Info(ammSwapTransaction.BuyTransaction)

	// Now we get the accounts we need.
	userBaseTokenAccount, _, err := solana.FindAssociatedTokenAddress(task.Wallet.PublicKey(), ammSwapTransaction.BuyTransaction.BaseMint)
	if err != nil {
		slog.Error("Error finding associated token address: " + err.Error())
		return
	}

	userQuoteTokenAccount, _, err := solana.FindAssociatedTokenAddress(task.Wallet.PublicKey(), ammSwapTransaction.BuyTransaction.QuoteMint)
	if err != nil {
		slog.Error("Error finding associated token address: " + err.Error())
		return
	}

	protocolFeeRecipient := pumpfunsdk.AMM_FEE_RECIPIENTS[rand.Intn(len(pumpfunsdk.AMM_FEE_RECIPIENTS))]
	protocolFeeRecipientTokenAccount, _, err := solana.FindAssociatedTokenAddress(protocolFeeRecipient, ammSwapTransaction.BuyTransaction.QuoteMint)
	if err != nil {
		slog.Error("Error finding associated token address: " + err.Error())
		return
	}

	// Now we print the accounts we need.
	slog.Info("📜 AMM Buy Transaction:")
	slog.Info(fmt.Sprintf("🔐 User Base Token Account: %s", userBaseTokenAccount.String()))
	slog.Info(fmt.Sprintf("🔐 User Quote Token Account: %s", userQuoteTokenAccount.String()))
	slog.Info(fmt.Sprintf("🔐 Protocol Fee Recipient: %s", protocolFeeRecipient.String()))
	slog.Info(fmt.Sprintf("🔐 Protocol Fee Recipient Token Account: %s", protocolFeeRecipientTokenAccount.String()))
	slog.Info("--------------------")

	tx, err = BuildBuyTransaction(ct, task, recentBlockHash, txSig, ammEvent.Buy, ammSwapTransaction, userBaseTokenAccount, userQuoteTokenAccount, protocolFeeRecipient, protocolFeeRecipientTokenAccount)
	if tx == nil {
		slog.Error("Error building buy transaction: tx nil")
		return
	}
	if err != nil {
		slog.Error("Error building buy transaction: " + err.Error())
		return
	}

	// print the transaction
	slog.Debug(spew.Sdump(tx))

	swapInfo := models.SwapInfo{
		License:              ct.UserConfig.LicenseKey,
		Timestamp:            utils.GetTimeString(),
		UserWallet:           task.Wallet.PublicKey().String(),
		ProgramName:          ct.GetProgramName(),
		IsBuy:                true,
		TokenMint:            ammSwapTransaction.BuyTransaction.BaseMint.String(),
		TokenMintAmount:      0,
		SolanaAmount:         task.SolBuyAmount / 1e9,
		TransactionSignature: tx.Signatures[0].String(),
	}
	go ct.db.LogSwap(&swapInfo)
	go utils.SendAnonWebhook(swapInfo)
	if ct.UserConfig.WebhookURL != "" {
		go utils.SendWebhook(ct.UserConfig.WebhookURL, swapInfo, false)
	}

	//// Send the transaction
	success, err := ct.CopyTraderConfig.Sender.SendTransaction(*tx, strategy.SendParams{
		TaskName:         task.TaskName,
		SimulationFailed: false,
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

	task.TaskMaximumBuysCounter[ammSwapTransaction.BuyTransaction.BaseMint.String()]++
	ct.CopyTraderConfig.GlobalBuyCounter[ammSwapTransaction.BuyTransaction.BaseMint.String()]++

	//os.Exit(0)
}

func HandleSell(ct *PfAmmCt, task *models.CopyTraderTask, recentBlockHash solana.Hash, txSig solana.Signature, ammEvent *pumpfunsdk.AMMEvent, tx *solana.Transaction, innerInstructions *[]rpc.InnerInstruction, balances utils.CustomPrePostTokenBalances) {
	// We know that we should sell the token in the amm event.
	ammSwapTransaction, err := pumpfunsdk.ParseAMMTransaction(tx, innerInstructions)
	if err != nil {
		slog.Error("Error parsing AMM transaction: " + err.Error())
		return
	}

	userBaseTokenAccount, _, err := solana.FindAssociatedTokenAddress(task.Wallet.PublicKey(), ammSwapTransaction.SellTransaction.BaseMint)
	if err != nil {
		slog.Error("Error finding associated token address: " + err.Error())
		return
	}

	protocolFeeRecipient := pumpfunsdk.AMM_FEE_RECIPIENTS[rand.Intn(len(pumpfunsdk.AMM_FEE_RECIPIENTS))]
	protocolFeeRecipientTokenAccount, _, err := solana.FindAssociatedTokenAddress(protocolFeeRecipient, ammSwapTransaction.SellTransaction.QuoteMint)
	if err != nil {
		slog.Error("Error finding associated token address: " + err.Error())
		return
	}

	decimals := balances.GetDecimals(ammSwapTransaction.SellTransaction.BaseMint.String())
	tx, err = BuildSellTransaction(ct, task, recentBlockHash, txSig, ammEvent.Sell, ammSwapTransaction, userBaseTokenAccount, protocolFeeRecipient, protocolFeeRecipientTokenAccount, int(decimals))
	if tx == nil {
		slog.Error("Error building buy transaction: tx nil")
		return
	}
	if err != nil {
		slog.Error("Error building buy transaction: " + err.Error())
		return
	}

	slog.Debug(spew.Sdump(tx))

	swapInfo := models.SwapInfo{
		License:              ct.UserConfig.LicenseKey,
		Timestamp:            utils.GetTimeString(),
		UserWallet:           task.Wallet.PublicKey().String(),
		ProgramName:          ct.GetProgramName(),
		IsBuy:                false,
		TokenMint:            ammSwapTransaction.SellTransaction.BaseMint.String(),
		TokenMintAmount:      0,
		SolanaAmount:         task.SolBuyAmount,
		TransactionSignature: tx.Signatures[0].String(),
	}
	go ct.db.LogSwap(&swapInfo)
	go utils.SendAnonWebhook(swapInfo)
	if ct.UserConfig.WebhookURL != "" {
		go utils.SendWebhook(ct.UserConfig.WebhookURL, swapInfo, false)
	}

	success, err := ct.CopyTraderConfig.Sender.SendTransaction(*tx, strategy.SendParams{
		TaskName:         task.TaskName,
		SimulationFailed: false,
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

	//Check if exists first , before decrementing
	x, ok := task.TaskMaximumBuysCounter[ammSwapTransaction.SellTransaction.BaseMint.String()]
	if ok {
		if x > 0 {
			task.TaskMaximumBuysCounter[ammSwapTransaction.SellTransaction.BaseMint.String()]--
		}
	}
	y, ok := ct.CopyTraderConfig.GlobalBuyCounter[ammSwapTransaction.SellTransaction.BaseMint.String()]
	if ok {
		if y > 0 {
			ct.CopyTraderConfig.GlobalBuyCounter[ammSwapTransaction.SellTransaction.BaseMint.String()]--
		}
	}

}
