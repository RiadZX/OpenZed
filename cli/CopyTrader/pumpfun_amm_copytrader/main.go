package pumpfun_amm_copytrader

import (
	"Zed/db"
	"Zed/models"
	"Zed/pumpfunsdk"
	"Zed/strategy"
	strategy2 "Zed/strategy"
	"Zed/utils"
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

type PfAmmCt struct {
	initialized      bool
	UserConfig       *models.Config
	CopyTraderConfig *models.CopyTraderConfig
	db               *db.Connection
}

func NewPfAmmCt(userConfig *models.Config, copyTraderConfig *models.CopyTraderConfig) (PfAmmCt, error) {
	if userConfig == nil || copyTraderConfig == nil {
		return PfAmmCt{}, errors.New("userConfig or copyTraderConfig is nil")
	}
	return PfAmmCt{
		UserConfig:       userConfig,
		CopyTraderConfig: copyTraderConfig,
	}, nil
}

func (ct *PfAmmCt) Init() error {
	if ct.UserConfig == nil {
		return errors.New("config not available")
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
	//initialize the cache
	ct.initialized = true
	return nil
}
func (ct *PfAmmCt) HandleGRPCMessage(geyserTx *geyser_pb.SubscribeUpdateTransaction, blockhash *solana.Hash) {
	if geyserTx == nil || blockhash == nil {
		return
	}
	if geyserTx.Transaction.Meta.Err != nil {
		sig := solana.SignatureFromBytes(geyserTx.Transaction.Signature)
		strategy.SigBuffer.SetStatusAndLog(sig.String(), strategy.TransactionError)
		slog.Debug("PumpFun AMM | Error in transaction | Sig: " + solana.SignatureFromBytes(geyserTx.Transaction.Signature).String())
		slog.Debug(spew.Sdump(geyserTx.Transaction.Meta.Err))
		return
	}
	slog.Info("PumpFun AMM | Handling gRPC message | Sig: " + solana.SignatureFromBytes(geyserTx.Transaction.Signature).String())
	transactionType, ammEventData := pumpfunsdk.ParseAMMLog(geyserTx.Transaction.Meta.LogMessages)
	if ammEventData == nil || (transactionType != pumpfunsdk.AMMSell && transactionType != pumpfunsdk.AMMBuy) {
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
			go Handle(ct, task, *blockhash, parsedTx.Signatures[0], ammEventData, parsedTx, &innerInstructions, balances)
		}
		if parsedTx.IsSigner(solana.MustPublicKeyFromBase58(task.Wallet.PublicKey().String())) {
			strategy2.SigBuffer.SetStatusAndLog(parsedTx.Signatures[0].String(), strategy2.TransactionSuccess)
		}
	}
}
func (ct *PfAmmCt) Start() error {
	if !ct.initialized {
		return errors.New("copyTrader not initialized")
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
func (ct *PfAmmCt) HandleWSMessage(got *ws.LogResult, blockhash *solana.Hash) {
	if got == nil || blockhash == nil {
		return
	}
	// TODO implement me
	panic("implement me")
}
func (ct *PfAmmCt) GetProgramName() string {
	return "PumpFun AMM"
}
func (ct *PfAmmCt) SetDatabase(db *db.Connection) {
	ct.db = db
}
