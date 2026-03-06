package models

import (
	strategy "Zed/strategy/transaction"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/weeaa/jito-go/clients/searcher_client"
	"os"
)

// CopyTraderConfig holds the configuration for the copy trading bot.
type CopyTraderConfig struct {
	HeliusUrl string `toml:"heliusUrl"`

	GetTransactionDelayMS int64                   `toml:"getTransactionDelayMS"`
	GetTransactionRetries int64                   `toml:"getTransactionRetries"`
	GetAccountInfoRetries int64                   `toml:"getAccountInfoRetries"`
	GetAccountInfoDelayMS int64                   `toml:"getAccountInfoDelayMS"`
	Jito                  bool                    `toml:"jito"`
	Circular              bool                    `toml:"circular"`
	Nextblock             bool                    `toml:"nextblock"`
	NextblockLocation     string                  `toml:"nextblockLocation"`
	NextblockAPIKEY       string                  `toml:"nextblock_apiKey"`
	BloxRoute             bool                    `toml:"bloxroute"`
	BloxRouteURL          string                  `toml:"bloxroute_url"`
	BloxRouteAPIKey       string                  `toml:"bloxroute_api_key"`
	BuyTipFee             float64                 `toml:"buy_tipFee"`
	SellTipFee            float64                 `toml:"sell_tipFee"`
	JitoSearcherClient    *searcher_client.Client `toml:"-"`
	AllTasks              []CopyTraderTask
	ActiveTasks           []*CopyTraderTask //these are all the tasks that are enabled

	// GlobalBuyCounter : this works with the  CopyTraderConfig.MaximumBuysPerToken
	// For every buy -> +1
	// For every sell -> -1
	// If limit has reached -> can not buy anymore
	// else -> Buy
	GlobalBuyCounter    map[string]int64
	MaximumBuysPerToken int64 `toml:"global_maximum_buys_per_token"`

	// Sending Strategy
	Sender         strategy.SendStrategy `toml:"-"`
	TemporalURL    string                `toml:"temporalLocation"`
	TemporalAPIKey string                `toml:"temporal_apiKey"`
	Temporal       bool                  `toml:"temporal"`
	CircularAPIKey string                `toml:"circular_apiKey"`

	// Turbo
	Turbo bool `toml:"turbo"`
}

// CopyTraderTask represents a single copy trading task.
type CopyTraderTask struct {
	Enabled      bool   `csv:"enabled"`
	TaskName     string `csv:"taskname"`
	PrivateKey   string `csv:"privatekey"`
	WalletToCopy string `csv:"wallet_to_copy"`

	StopLoss   float64 `csv:"stop_loss"`
	TakeProfit float64 `csv:"take_profit"`

	DynamicFee      bool    `csv:"dynamic_fee"`
	DynamicFeeLevel string  `csv:"dynamic_fee_level"`
	FixedFee        float64 `csv:"fixed_fee"`
	SolBuyAmount    float64 `csv:"sol_buy_amount"`
	BuySlippage     float64 `csv:"buy_slippage"`
	SellingSlippage float64 `csv:"selling_slippage"`
	MonitorMode     bool    `csv:"monitormode"`
	// BuyPercentage is the percentage of the wallet to copy that will be used to buy tokens.
	// If it is set to 0 -> the SolBuyAmount will be used. Otherwise a percentage of the wallet will be used.
	BuyPercentage      float64 `csv:"buy_percentage"`
	CopySellPercentage bool    `csv:"copy_sell_percentage"`
	CopySell           bool    `csv:"copy_sell"`
	MinimumMarket      int     `csv:"min_mcap"`

	TaskMaximumBuys        int64 `csv:"task_maximum_buys_per_token"`
	TaskMaximumBuysCounter map[string]int64

	FirstTimeBuyCache FirstTimeBuyCache    `csv:"-"`
	TokensOwnedCache  TokensOwnedCache     `csv:"-"`
	Status            CopyTraderTaskStatus `csv:"-"`
	Wallet            *solana.Wallet       `csv:"-"` // this field is not in the csv file

}

func (c *Config) SaveCopyTraderTasks(filename string) error {
	//save back to tasks.csv file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)

		}
	}(file)

	tasksString := tasksToCSVString(c.CopyTraderConfig.AllTasks)
	_, err = file.WriteString(tasksString)
	if err != nil {
		return err
	}
	return nil
}

func tasksToCSVString(tasks []CopyTraderTask) string {
	//header
	csvString := "taskname,privatekey,wallet_to_copy,dynamic_fee,dynamic_fee_level,fixed_fee,sol_buy_amount,buy_slippage,selling_slippage,task_maximum_buys_per_token,copy_sell_percentage,monitormode,enabled\n"
	for i := 0; i < len(tasks); i++ {
		task := &tasks[i]
		csvString += fmt.Sprintf("%s,%s,%s,%t,%s,%f,%f,%f,%f,%d,%t,%t,%t\n", task.TaskName, task.PrivateKey, task.WalletToCopy, task.DynamicFee, task.DynamicFeeLevel, task.FixedFee, task.SolBuyAmount, task.BuySlippage, task.SellingSlippage, task.TaskMaximumBuys, task.CopySellPercentage, task.MonitorMode, task.Enabled)
	}
	return csvString
}

type CopyTraderTaskStatus struct {
	TaskName            string
	BuysDetected        int
	SellsDetected       int
	SentTransactions    int
	SuccessTransactions int
	FailedTransactions  int
	PendingTransactions int
	PumpVisionLink      string
}

func (c *CopyTraderTaskStatus) String() string {
	return fmt.Sprintf("TaskName: %s, BuysDetected: %d, SellsDetected: %d, SentTransactions: %d, SuccessTransactions: %d, FailedTransactions: %d, PendingTransactions: %d, PumpVisionLink: %s",
		c.TaskName, c.BuysDetected, c.SellsDetected, c.SentTransactions, c.SuccessTransactions, c.FailedTransactions, c.PendingTransactions, c.PumpVisionLink)
}
