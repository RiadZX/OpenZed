package pumpfuncopytrader

import (
	"Zed/db"
	"Zed/models"
	"Zed/pumpfunsdk"
	"Zed/strategy"
	"Zed/triggers"
	"Zed/utils"
	"context"
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gookit/slog"
	"github.com/weeaa/goyser"
	geyser_pb "github.com/weeaa/goyser/pb"
	"net/url"
)

type PFCopyTrader struct {
	initialized      bool
	UserConfig       *models.Config
	CopyTraderConfig *models.CopyTraderConfig
	db               *db.Connection
}

func (ct *PFCopyTrader) SetDatabase(db *db.Connection) {
	ct.db = db
}

func (ct *PFCopyTrader) GetProgramName() string {
	return "PumpFun"
}

func (ct *PFCopyTrader) HandleGRPCMessage(geyserTx *geyser_pb.SubscribeUpdateTransaction, blockhash *solana.Hash) {
	if geyserTx == nil || blockhash == nil {
		return
	}
	if geyserTx.Transaction.Meta.Err != nil {
		sig := solana.SignatureFromBytes(geyserTx.Transaction.Signature)
		strategy.SigBuffer.SetStatusAndLog(sig.String(), strategy.TransactionError)
		slog.Debug("PumpFun | Error in transaction | Sig: " + solana.SignatureFromBytes(geyserTx.Transaction.Signature).String())
		slog.Debug(spew.Sdump(geyserTx.Transaction.Meta.Err))
		return
	}
	slog.Info("PumpFun | Handling gRPC message | Sig: " + solana.SignatureFromBytes(geyserTx.Transaction.Signature).String())
	shouldProcess, transactionType, tradeData := pumpfunsdk.ParseLog(geyserTx.Transaction.Meta.LogMessages)
	if !shouldProcess || tradeData == nil || (transactionType != pumpfunsdk.Buy && transactionType != pumpfunsdk.Sell) {
		return
	}
	parsedTx := goyser.ConvertTransaction(geyserTx)
	slog.Info("PumpFun | Parsed transaction")
	innerInstructions := utils.ParseInnerInstructions(geyserTx.Transaction.Meta.InnerInstructions)

	balances := utils.CustomPrePostTokenBalances{
		PreTokenBalances:  []utils.XTokenBalance{},
		PostTokenBalances: []utils.XTokenBalance{},
	}

	err := balances.FromGeyserPrePro(geyserTx.Transaction.Meta.PreTokenBalances, geyserTx.Transaction.Meta.PostTokenBalances)
	if err != nil {
		slog.Error("error getting balances: " + err.Error())
		return
	}
	for _, task := range ct.CopyTraderConfig.ActiveTasks {
		if parsedTx.IsSigner(solana.MustPublicKeyFromBase58(task.WalletToCopy)) {
			go HandleBuySellTransaction(ct, task, *blockhash, parsedTx.Signatures[0], tradeData, parsedTx, &innerInstructions, balances)
		}
		//check first if this signature is needed for us. if it was self-made.
		if parsedTx.IsSigner(solana.MustPublicKeyFromBase58(task.Wallet.PublicKey().String())) {
			strategy.SigBuffer.SetStatusAndLog(parsedTx.Signatures[0].String(), strategy.TransactionSuccess)
			if task.StopLoss == 0 || task.TakeProfit == 0 {
				slog.Info("PumpFun | Stoploss or Takeprofit is 0, skipping")
				continue
			}
			//// Add Trigger for Stoploss/take profit.
			_, txType, tradeEventData := pumpfunsdk.ParseLog(geyserTx.Transaction.Meta.LogMessages)
			if txType != pumpfunsdk.Buy {
				continue
			}
			bondingCurve := &pumpfunsdk.BondingCurve{
				VirtualTokenReserves: tradeEventData.VirtualTokenReserves,
				VirtualSolReserves:   tradeEventData.VirtualSolReserves,
				RealTokenReserves:    tradeEventData.RealTokenReserves,
				RealSolReserves:      tradeEventData.RealSolReserves,
				TokenTotalSupply:     1000000000000000,
				Complete:             false,
			}
			price, err := pumpfunsdk.CalculateTokenPrice(bondingCurve)
			if err != nil {
				slog.Error("Error getting token price: " + err.Error())
				continue
			}
			fac := triggers.PumpfunTradeFactory{}
			trade := fac.CreateTrade(tradeData.Mint.String(), price, task.StopLoss, task.TakeProfit, triggers.ExecutionOpts{
				Wallet:     *task.Wallet,
				FixedFee:   task.FixedFee,
				Signature:  parsedTx.Signatures[0],
				Sender:     ct.CopyTraderConfig.Sender,
				TokenCache: &task.TokensOwnedCache,
				Blockhash:  blockhash,
			})
			slog.Info("PumpFun | Created trade: " + trade.GetMint())
			err = triggers.MonitorInstance.AddTrade(trade)
			if err != nil {
				slog.Error("Error adding trade to monitor: " + err.Error())
				return
			}

		}
	}

}

func NewPFCopyTrader(userConfig *models.Config, copyTraderConfig *models.CopyTraderConfig) (PFCopyTrader, error) {
	if userConfig == nil || copyTraderConfig == nil {
		return PFCopyTrader{}, errors.New("userConfig or copyTraderConfig is nil")
	}
	return PFCopyTrader{
		UserConfig:       userConfig,
		CopyTraderConfig: copyTraderConfig,
	}, nil
}

func (ct *PFCopyTrader) Init() error {
	if ct.UserConfig == nil {
		return ErrConfigNotAvailable
	}
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

	//initialize the statuses
	//	err = ct.initStatuses()
	//if err != nil {
	//	return err
	//	}
	//initialize the cache
	ct.initialized = true
	return nil
}

func (ct *PFCopyTrader) Start() error {
	if !ct.initialized {
		return ErrNotInitialized
	}
	slog.Info("PumpFun | Initializing cache")
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

	slog.Info("Cache has been initialized")
	return nil
}

func (ct *PFCopyTrader) HandleWSMessage(got *ws.LogResult, blockhash *solana.Hash) {
	if got == nil || blockhash == nil {
		return
	}

	if got.Value.Err != nil {
		strategy.SigBuffer.SetStatusAndLog(got.Value.Signature.String(), strategy.TransactionError)
		//slog.Debug("PumpFun | Error in transaction | Sig: " + got.Value.Signature.String())
		return
	}
	shouldProcess, transactionType, tradeData := pumpfunsdk.ParseLog(got.Value.Logs)
	if !shouldProcess || tradeData == nil || (transactionType != pumpfunsdk.Buy && transactionType != pumpfunsdk.Sell) {
		//if transactionType == Error && tradeData == nil {
		//	slog.Error("This shit be buggging , sig " + got.Value.Signature.String())
		//}
		return
	}
	//check if user is someone we care about
	pass := false
	for _, task := range ct.CopyTraderConfig.ActiveTasks {
		if tradeData.User.String() == task.WalletToCopy || tradeData.User.String() == task.Wallet.PublicKey().String() {
			pass = true
			break
		}
	}
	if !pass {
		return
	}
	//*API CALL
	transactionData, err := utils.GetTransaction(context.Background(), ct.UserConfig.HttpClient, got.Value.Signature.String(), ct.UserConfig.ExtraRPCS, ct.CopyTraderConfig.GetTransactionDelayMS, ct.CopyTraderConfig.GetTransactionRetries)
	if err != nil || transactionData == nil {
		if err != nil {
			slog.Error(err.Error())
		}

		return
	}

	tx, err := transactionData.Transaction.GetTransaction()
	if err != nil {
		slog.Error("Error: " + err.Error() + "Sig: " + got.Value.Signature.String())
		return
	}
	innerInstructions := transactionData.Meta.InnerInstructions

	balances := utils.CustomPrePostTokenBalances{
		PreTokenBalances:  []utils.XTokenBalance{},
		PostTokenBalances: []utils.XTokenBalance{},
	}
	err = balances.FromSolanaPrePro(transactionData.Meta.PreTokenBalances, transactionData.Meta.PostTokenBalances)
	if err != nil {
		slog.Error("error getting balances: " + err.Error())
		return
	}

	for _, task := range ct.CopyTraderConfig.ActiveTasks {
		if tx.IsSigner(solana.MustPublicKeyFromBase58(task.WalletToCopy)) {
			HandleBuySellTransaction(ct, task, *blockhash, got.Value.Signature, tradeData, tx, &innerInstructions, balances)
		}
		if tx.IsSigner(solana.MustPublicKeyFromBase58(task.Wallet.PublicKey().String())) {
			strategy.SigBuffer.SetStatusAndLog(tx.Signatures[0].String(), strategy.TransactionSuccess)
			if task.StopLoss == 0 || task.TakeProfit == 0 {
				slog.Info("PumpFun | Stoploss or Takeprofit is 0, skipping")
				continue
			}
			//// Add Trigger for Stoploss/take profit.
			_, txType, tradeEventData := pumpfunsdk.ParseLog(transactionData.Meta.LogMessages)
			if txType != pumpfunsdk.Buy {
				continue
			}
			bondingCurve := &pumpfunsdk.BondingCurve{
				VirtualTokenReserves: tradeEventData.VirtualTokenReserves,
				VirtualSolReserves:   tradeEventData.VirtualSolReserves,
				RealTokenReserves:    tradeEventData.RealTokenReserves,
				RealSolReserves:      tradeEventData.RealSolReserves,
				TokenTotalSupply:     1000000000000000,
				Complete:             false,
			}
			price, err := pumpfunsdk.CalculateTokenPrice(bondingCurve)
			if err != nil {
				slog.Error("Error getting token price: " + err.Error())
				continue
			}
			fac := triggers.PumpfunTradeFactory{}
			trade := fac.CreateTrade(tradeData.Mint.String(), price, task.StopLoss, task.TakeProfit, triggers.ExecutionOpts{
				Signature:  tx.Signatures[0],
				Wallet:     *task.Wallet,
				FixedFee:   task.FixedFee,
				Sender:     ct.CopyTraderConfig.Sender,
				TokenCache: &task.TokensOwnedCache,
				Blockhash:  blockhash,
			})
			slog.Info("PumpFun | Created trade: " + trade.GetMint())
			err = triggers.MonitorInstance.AddTrade(trade)
			if err != nil {
				slog.Error("Error adding trade to monitor: " + err.Error())
				return
			}

		}
	}

}
