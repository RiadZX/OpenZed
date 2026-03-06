package pumpfuncopytrader

import (
	"Zed/models"
	"Zed/pumpfunsdk"
	"Zed/utils"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gookit/slog"
	"math"
)

func getSellIX(ct *PFCopyTrader, task *models.CopyTraderTask, txSig solana.Signature, ata string, p *PurchaseSellData, curve *pumpfunsdk.BondingCurve) ([]solana.Instruction, float64, float64, error) {
	//make sure that we have the token to sell, so check cache, if not in cache then call to see if we have the token
	mint := p.Mint.String()
	err := task.TokensOwnedCache.Preload(ct.UserConfig.HttpClient, task.Wallet.PublicKey().String())
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", txSig)
		return nil, 0, 0, err
	}
	if !task.TokensOwnedCache.Exists(mint) {
		return nil, 0, 0, fmt.Errorf("token not owned")
	}
	tokenAmount, exists := task.TokensOwnedCache.Get(mint) //this is the amount of tokens we have 100%, this is the UI Amount
	if !exists {
		return nil, 0, 0, fmt.Errorf("token not owned")
	}

	//Sell in according to the percentage of the other wallet.
	if task.CopySellPercentage {
		slog.Debug("Starting getting the percentage")

		percentage, err := GetSellPercentage(p.PreTokenBalance, p.PostTokenBalance)
		if err != nil {
			slog.Error("Error getting percentage: " + err.Error())
			return nil, 0, 0, err
		}
		//use the same percentage to sell the tokens
		tokenAmount = tokenAmount * percentage
		slog.Debug("Percentage: ", percentage)
		slog.Debug("Done getting the percentage")
	}

	pricePerTokens, err := pumpfunsdk.CalculateTokenPrice(curve)
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", txSig)
		return nil, 0, 0, err
	}
	price := tokenAmount * pricePerTokens
	//factor in slippage
	price = price * (1 - task.SellingSlippage)
	//construct the sell instruction
	var sellIX solana.Instruction

	sellIX, err = pumpfunsdk.ConstructSellInstruction(
		uint64(tokenAmount*1e6),
		uint64(price),
		p.Mint,
		solana.MustPublicKeyFromBase58(p.BondingCurve),
		solana.MustPublicKeyFromBase58(p.AssociatedBondingCurve),
		solana.MustPublicKeyFromBase58(ata),
		task.Wallet.PublicKey(),
	)
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", txSig)
		return nil, 0, 0, err
	}

	return []solana.Instruction{sellIX}, tokenAmount, price, nil
}

// GetSellPercentage tokensSold is the amount of tokens the wallet has told, this is in decimals so already multiplied by 1e6
func GetSellPercentage(preTokenBalance, postTokenBalance float64) (float64, error) {
	if postTokenBalance == 0 {
		return 1, nil
	}

	percentage := (preTokenBalance - postTokenBalance) / preTokenBalance

	if percentage > 1.01 {
		return 0, fmt.Errorf("math error")
	}
	return percentage, nil
}

func getBuyIX(task *models.CopyTraderTask, txSig solana.Signature, ata string, purchaseData *PurchaseSellData, bondingCurve *pumpfunsdk.BondingCurve) (tx []solana.Instruction, tokenamount float64, price float64, err error) {

	// Check if percentage purchase is enabled or not
	var solBuyAmount = task.SolBuyAmount
	if task.BuyPercentage > 0 {
		//task.BuyPercentage is (0, 1]
		// max is set to the max amount of sol that can be used to buy tokens
		solBuyAmount = math.Min(task.BuyPercentage*float64(purchaseData.SolAmount)/1e9, task.SolBuyAmount)
	}
	tokenAmount := pumpfunsdk.CalculateTokensToBuy(bondingCurve, solBuyAmount)
	if tokenAmount == 0 {
		slog.Warn("Amount you set in was too low, brokie")
		return nil, 0, 0, fmt.Errorf("sol amount was too low")
	}

	isFirstTime, err := isFirstTimeBuyingCache(ata, &task.FirstTimeBuyCache)
	if err != nil {
		slog.Error("Error: ", err, "Sig: ", txSig)
		return nil, 0, 0, err
	}

	var createATAIX solana.Instruction
	var buyIX solana.Instruction

	if isFirstTime {
		createATAIX, err = utils.ConstructCreateATAInstruction(task.Wallet.PublicKey().String(), purchaseData.Mint.String())
		if err != nil {
			slog.Error("Error: ", err, "Sig: ", txSig)
			return nil, 0, 0, err
		}
		if createATAIX == nil {
			slog.Error("Error: ", err, "Sig: ", txSig)
			return nil, 0, 0, fmt.Errorf("createATAIX is nil")
		}
	}
	maxSolAmount := solBuyAmount * (1 + task.BuySlippage) * 1e9 //convert to lamports
	buyIX, err = pumpfunsdk.ConstructBuyInstruction(
		task.Wallet.PublicKey(),
		tokenAmount,
		uint64(maxSolAmount),
		purchaseData.Mint,
		solana.MustPublicKeyFromBase58(purchaseData.BondingCurve),
		solana.MustPublicKeyFromBase58(purchaseData.AssociatedBondingCurve),
		solana.MustPublicKeyFromBase58(ata),
	)

	if err != nil {
		slog.Error("Error: ", err, "Sig: ", txSig)
		return nil, 0, 0, err
	}

	if isFirstTime {
		return []solana.Instruction{createATAIX, buyIX}, float64(tokenAmount), maxSolAmount, nil
	} else {
		return []solana.Instruction{buyIX}, float64(tokenAmount), maxSolAmount, nil
	}
}
