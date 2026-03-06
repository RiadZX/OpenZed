package raydiumcopytrader

import (
	"Zed/mocks"
	"Zed/models"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gookit/slog"
	"math/big"
	"testing"
)

func TestMarketCapComparison(t *testing.T) {
	// Set up test values
	pricePerToken := int64(81043)
	supply := int64(22935105504941)
	mintDecimals := int64(6)
	solPrice := int64(167)
	minimumMarketCap := int64(3200)

	// Convert to big.Int
	pricePerTokenBig := big.NewInt(pricePerToken)
	supplyBig := big.NewInt(supply)
	mintDecimalsBig := big.NewInt(mintDecimals)
	solPriceBig := big.NewInt(solPrice)
	minimumMarketCapBig := big.NewInt(minimumMarketCap)

	// Calculate (supply / 10^mintDecimals)
	supplyDiv := new(big.Int).Div(supplyBig, new(big.Int).Exp(big.NewInt(10), mintDecimalsBig, nil))

	// Calculate (pricePerToken * supplyDiv * solPrice)
	marketCap := new(big.Int).Mul(pricePerTokenBig, supplyDiv)
	marketCap.Mul(marketCap, solPriceBig)

	// Perform the comparison
	if marketCap.Cmp(minimumMarketCapBig) < 0 {
		slog.Info("Marketcap too low, not sending transaction")
	} else {
		slog.Info("Marketcap is high enough, sending transaction")
	}
}

var file = `
licenseKey = "9e7ed783-942e-4c6b-8dd6-ed47b54719a0"
#9e7ed783-942e-4c6b-8dd6-ed47b54719a0 = zed
#8fab32a7-cf39-4ba5-a25d-2b45bbc365ef
rpcUrl = "https://red-old-flower.solana-mainnet.quiknode.pro/8edd9abee6d707afffd059022e1ece558208243d"
grpcUrl = "http://newyork.omeganetworks.io:10000"
wssUrl = "ws://newyork.omeganetworks.io:8900"
webhookUrl="https://discord.com/api/webhooks/WEBHOOKURL"
blockEngineURLLocation="amsterdam" #amsterdam,frankfurt,newyork,tokyo
TPS = 5 #transactions per second
#ONLY USED WHEN NOT USING gRPC. IF USING GRPC THEN IGNORE.
extraRPCS=[ #these will be used for getting the transaction details, multiple to counter rate limit
]
pnlDelayMS = 500
grpcHeaders = ""
[copyTrader]
# maximum amount of times the bot is allowed to buy of a token
global_maximum_buys_per_token = 1000
turbo=true
## -------------------
temporal = false #use temporal to send
jito = false #use jito to send
nextblock = false #use nextblock to send
circular = false #use circular to send
temporalLocation = "amsterdam" #amsterdam / frankfurt / useast
nextblockLocation = "frankfurt" # frankfurt / newyork
## -------------------
tipFee = 0.001
heliusUrl="https://mainnet.helius-rpc.com/?api-key=XYZ"
getTransactionDelayMS = 500 #try and change this
getTransactionRetries = 50 #leave at 50!
getAccountInfoDelayMS = 400
getAccountInfoRetries = 50
`
var tasks = `
taskname,privatekey,wallet_to_copy,stop_loss,take_profit,dynamic_fee,dynamic_fee_level,fixed_fee,sol_buy_amount,buy_slippage,selling_slippage,task_maximum_buys_per_token,copy_sell_percentage,copy_sell,min_mcap,monitormode,enabled
ray,373QdDxPvSaByTQ3pEPCRmFjAy8gVYKzU8yrA3TgUirHh45uz5t4ArGm9VpSD8hcQNqSh1FABxfE7iuutnWskDpk,2FZczWjKj1JVAn2DWr2WL5V9VzvPtf2YRQ8h3PWPj4aW,0.95,1.25,false,medium,0.01,0.01,0.5,1,1000,true,false,0,true,true`

func Test_HandleWSMessage(t *testing.T) {
	// Capture log output
	logBuffer := mocks.CaptureLogOutput()

	sig, err := solana.SignatureFromBase58("3NsqmmtFfpbumYLoXAso1NAmPgKjzPVRjCbsWutoAtoyzi4WbGtjhqJFin1DMPhaoq4BYfTHk6HRA4STLRovRhcn")
	if err != nil {
		t.Errorf("HandleWSMessage() error = %v", err)
	}
	got := ws.LogResult{
		Value: struct {
			Signature solana.Signature `json:"signature"`
			Err       interface{}      `json:"err"`
			Logs      []string         `json:"logs"`
		}{
			Signature: sig,
			Err:       nil,
			Logs:      testLogsWSMEssage,
		},
	}

	conf, err := models.NewConfig(file, "Zed")
	if err != nil {
		t.Fatal(err)
	}
	err = conf.NewTasks(tasks)
	if err != nil {
		t.Fatal(err)
	}

	ct, err := NewRaydiumCopyTrader(&conf, conf.CopyTraderConfig)
	if err != nil {
		t.Fatal(err)
	}

	ct.HandleWSMessage(&got, &solana.Hash{})
	// Print captured log output
	t.Log(logBuffer.String())
}

var testLogsWSMEssage = []string{
	"Program 11111111111111111111111111111111 invoke [1]",
	"Program 11111111111111111111111111111111 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: InitializeAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3443 of 799850 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 invoke [1]",
	"Program log: ray_log: AyEr5zwAAAAAEq4BAAAAAAACAAAAAAAAAIUMfB8FAAAAKgPVyT41AADOPPDNlAAAAOvGqQAAAAAA",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: Transfer",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 777525 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: Transfer",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4736 of 769899 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 consumed 32322 of 796407 compute units",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: CloseAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2915 of 764085 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
}
