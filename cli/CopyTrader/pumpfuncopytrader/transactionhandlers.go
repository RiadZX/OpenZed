package pumpfuncopytrader

import (
	"Zed/models"
	"Zed/pumpfunsdk"
	strategy "Zed/strategy/transaction"
	"Zed/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"math/big"
)

func HandleBuySellTransaction(ct *PFCopyTrader, task *models.CopyTraderTask, recentBlockHash solana.Hash, txSig solana.Signature, tradeData *pumpfunsdk.TradeEventData, tx *solana.Transaction, innerInstructions *[]rpc.InnerInstruction, balances utils.CustomPrePostTokenBalances) {
	if tradeData == nil || tx == nil || innerInstructions == nil {
		slog.Error("Error: tradeData, tx or innerInstructions is nil")
		return
	}

	purchaseData, err := getPurchaseData(tradeData, tx, innerInstructions)
	if err != nil {
		slog.Warn("PumpFun | Could not get purchase data: " + err.Error())
		slog.Warn(spew.Sdump(tradeData))
		return
	}

	if !task.CopySell && !purchaseData.IsBuy {
		slog.Info("PumpFun | Sell detected, but not copying sells: " + txSig.String())
		return
	}

	purchaseData.PreTokenBalance = balances.GetPreTokenBalance(task.WalletToCopy, purchaseData.Mint.String())
	purchaseData.PostTokenBalance = balances.GetPostTokenBalance(task.WalletToCopy, purchaseData.Mint.String())

	if purchaseData.PreTokenBalance == 0 && purchaseData.PostTokenBalance == 0 && !purchaseData.IsBuy {
		slog.Error("Could not fetch the pre/post token balances: " + txSig.String())
		if task.CopySellPercentage {
			slog.Error("This is needed for copy % mode, so cancelling this signature, please report: " + txSig.String())
			return
		}
	}

	minimumMarketCap := task.MinimumMarket
	if minimumMarketCap != 0 && purchaseData.IsBuy {
		solPriceBig := big.NewInt(int64(utils.GetSolPrice()))
		virtualSolReservesBig := big.NewInt(int64(purchaseData.VirtualSolReserves))
		realSolReservesBig := big.NewInt(int64(purchaseData.RealSolReserves))
		minimumMarketCapBig := big.NewInt(int64(minimumMarketCap))

		// Calculate currentSolFloating = VirtualSolReserves + RealSolReserves
		currentSolFloatingBig := new(big.Int).Add(virtualSolReservesBig, realSolReservesBig)

		// Calculate currentMarketCap = (currentSolFloating * solPrice) / 1e9
		currentMarketCapBig := new(big.Int).Mul(currentSolFloatingBig, solPriceBig)
		currentMarketCapBig.Div(currentMarketCapBig, big.NewInt(1e9))

		//DEBUG LOG
		slog.Debug("SolPrice: ", solPriceBig)
		slog.Debug("VirtualSolReserves: ", virtualSolReservesBig)
		slog.Debug("RealSolReserves: ", realSolReservesBig)
		slog.Debug("MinimumMarketCap: ", minimumMarketCapBig)
		slog.Debug("CurrentMarketCap: ", currentMarketCapBig)
		slog.Debug("CurrentSolFloating: ", currentSolFloatingBig)

		if currentMarketCapBig.Cmp(minimumMarketCapBig) < 0 {
			slog.Warn("Marketcap is below the minimum: " + txSig.String())
			return
		}
	}

	go HandleTask(ct, task, purchaseData, txSig, recentBlockHash)
}

func HandleTask(ct *PFCopyTrader, task *models.CopyTraderTask, purchaseData PurchaseSellData, txSig solana.Signature, recentBlockHash solana.Hash) {
	if purchaseData.IsBuy {
		slog.Info(task.TaskName + " : detected a buy transaction. Sig: " + txSig.String())
		task.Status.BuysDetected++
	} else {
		slog.Info(task.TaskName + " : detected a sell transaction. Sig: " + txSig.String())
		task.Status.SellsDetected++
	}
	if task.MonitorMode {
		slog.Info(task.TaskName + " : Monitor mode enabled, not sending transaction. Sig: " + txSig.String())
		return
	}

	var commitment rpc.CommitmentType
	if ct.UserConfig.GRPCUrl != "" {
		commitment = rpc.CommitmentProcessed
	} else {
		commitment = rpc.CommitmentConfirmed
	}
	var bondingCurve *pumpfunsdk.BondingCurve
	var err error
	if ct.CopyTraderConfig.Turbo == false || purchaseData.SolAmount == 0 && purchaseData.VirtualTokenReserves == 0 {
		slog.Info("Using bonding curve for bonding curve")
		bondingCurve, err = pumpfunsdk.GetBondingCurve(ct.UserConfig.HttpClient, solana.MustPublicKeyFromBase58(purchaseData.BondingCurve), pumpfunsdk.ClientOpts{
			Commitment: commitment,
			Retries:    int(ct.CopyTraderConfig.GetAccountInfoRetries),
			Delay:      ct.CopyTraderConfig.GetAccountInfoDelayMS,
		})
		if err != nil {
			slog.Error("Error getting bonding curve:", err)
			return
		}
	} else {
		slog.Info("Using trade data for bonding curve")
		bondingCurve = &pumpfunsdk.BondingCurve{
			VirtualTokenReserves: purchaseData.VirtualTokenReserves,
			VirtualSolReserves:   purchaseData.VirtualSolReserves,
			RealTokenReserves:    purchaseData.RealTokenReserves,
			RealSolReserves:      purchaseData.RealSolReserves,
			TokenTotalSupply:     1000000000000000,
			Complete:             false,
		}
	}

	slog.Debug("Got bonding curve: ", bondingCurve)
	ata, err := utils.GetATA(task.Wallet.PublicKey().String(), purchaseData.Mint.String())
	if err != nil {
		slog.Error("Error: " + err.Error() + "Sig: " + txSig.String())
		return
	}
	slog.Debug("Got ATA: ", ata)

	//At this stage we should diverse to either sell or buy
	var instructions []solana.Instruction
	var tokenAmount, price float64
	switch purchaseData.IsBuy {
	case true:
		instructions, tokenAmount, price, err = getBuyIX(task, txSig, ata, &purchaseData, bondingCurve)
		slog.Debug("Got buy instructions")
		if err != nil {
			slog.Error("Error: " + err.Error() + "Sig: " + txSig.String())
			return
		}

		//Check if allowed to buy according to the task counter
		taskAmount, ok := task.TaskMaximumBuysCounter[purchaseData.Mint.String()]
		globalAmount, ok2 := ct.CopyTraderConfig.GlobalBuyCounter[purchaseData.Mint.String()]
		if !ok {
			taskAmount = 0
			task.TaskMaximumBuysCounter[purchaseData.Mint.String()] = 0
		}
		if !ok2 {
			globalAmount = 0
			ct.CopyTraderConfig.GlobalBuyCounter[purchaseData.Mint.String()] = 0
		}
		if taskAmount >= task.TaskMaximumBuys || globalAmount >= ct.CopyTraderConfig.MaximumBuysPerToken || taskAmount >= ct.CopyTraderConfig.MaximumBuysPerToken {
			slog.Info("Task: " + task.TaskName + " has reached the maximum buys for this token")
			return
		}

	case false:
		instructions, tokenAmount, price, err = getSellIX(ct, task, txSig, ata, &purchaseData, bondingCurve)
		slog.Debug("Got sell instructions")
		if err != nil {
			slog.Error("Error: " + err.Error() + "Sig: " + txSig.String())
			return
		}
	}

	optimalTx, err := utils.GetOptimalVersionedTransaction(ct.UserConfig.HttpClient, *ct.CopyTraderConfig, &recentBlockHash, purchaseData.IsBuy, task, ct.GetProgramName(), instructions...)
	slog.Debug("Transaction: ", spew.Sdump(optimalTx))
	slog.Debug("Error: ", spew.Sdump(err))
	if !utils.IsIgnorableError(err) {
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", optimalTx)
			return
		}
		return
	}
	slog.Debug("No errors in getting optimal transaction")

	swapInfo := models.SwapInfo{
		License:              ct.UserConfig.LicenseKey,
		Timestamp:            utils.GetTimeString(),
		UserWallet:           task.Wallet.PublicKey().String(),
		ProgramName:          ct.GetProgramName(),
		IsBuy:                purchaseData.IsBuy,
		TokenMint:            purchaseData.Mint.String(),
		TokenMintAmount:      tokenAmount,
		SolanaAmount:         price / 1e9,
		TransactionSignature: optimalTx.Signatures[0].String(),
	}
	go ct.db.LogSwap(&swapInfo)
	go utils.SendAnonWebhook(swapInfo)
	if ct.UserConfig.WebhookURL != "" {
		go utils.SendWebhook(ct.UserConfig.WebhookURL, swapInfo, false)
	}

	success, err := ct.CopyTraderConfig.Sender.SendTransaction(optimalTx, strategy.SendParams{
		TaskName:         task.TaskName,
		SimulationFailed: false,
	})
	if success {
		slog.Info("Successfully sent transaction: " + optimalTx.Signatures[0].String())
	} else {
		if err != nil {
			slog.Error("Unknown sending transaction: " + err.Error())
		} else {
			slog.Error("Unknown status sending transaction")
		}
	}

	models.RefreshCachesWithWallet(ct.UserConfig.HttpClient, task.Wallet.PublicKey().String(), ct.CopyTraderConfig)

	if purchaseData.IsBuy {
		task.TaskMaximumBuysCounter[purchaseData.Mint.String()]++
		ct.CopyTraderConfig.GlobalBuyCounter[purchaseData.Mint.String()]++
	} else {
		//Check if exists first , before decrementing
		x, ok := task.TaskMaximumBuysCounter[purchaseData.Mint.String()]
		if ok {
			if x > 0 {
				task.TaskMaximumBuysCounter[purchaseData.Mint.String()]--
			}
		}
		y, ok := ct.CopyTraderConfig.GlobalBuyCounter[purchaseData.Mint.String()]
		if ok {
			if y > 0 {
				ct.CopyTraderConfig.GlobalBuyCounter[purchaseData.Mint.String()]--
			}
		}
	}
}
