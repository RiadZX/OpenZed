package models

import (
	strategy "Zed/strategy/transaction"
	"context"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gocarina/gocsv"
	"github.com/gookit/slog"
	jito_go "github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/clients/searcher_client"
	"net/url"
	"os"
)

type Config struct {
	LicenseKey             string `toml:"licenseKey"`
	WebhookURL             string `toml:"webhookURL"`
	RPCUrl                 string `toml:"rpcUrl"`
	WSSUrl                 string `toml:"wssUrl"`
	GRPCUrl                string `toml:"grpcUrl"`
	BLockEngineURLLocation string `toml:"blockEngineURLLocation"`
	PnlDelayMS             int    `toml:"pnlDelayMS"`
	HttpClient             IClient
	CopyTraderConfig       *CopyTraderConfig `toml:"copyTrader"`
	ExtraRPCS              []string          `toml:"extraRPCS"`
	TPS                    int               `toml:"tps"`
	GRPCHeaders            string            `toml:"grpcHeaders"`
}

func NewConfig(data string, whitelabel string) (Config, error) {
	//data is the string of the toml file
	var config Config
	_, err := toml.Decode(data, &config)
	if err != nil {
		return Config{}, err
	}
	//validate rpc url and wss url
	_, err = url.ParseRequestURI(config.RPCUrl)
	if err != nil {
		slog.Error("Error parsing RPCUrl: " + err.Error())
		return Config{}, err
	}
	_, err = url.ParseRequestURI(config.WSSUrl)
	if err != nil {
		slog.Error("Error parsing WSSUrl: " + err.Error())
		return Config{}, err
	}

	config.HttpClient = rpc.New(config.RPCUrl)

	var blockEngineURl string
	switch config.BLockEngineURLLocation {
	case "amsterdam":
		blockEngineURl = jito_go.Amsterdam.BlockEngineURL
	case "frankfurt":
		blockEngineURl = jito_go.Frankfurt.BlockEngineURL
	case "newyork":
		blockEngineURl = jito_go.NewYork.BlockEngineURL
	case "tokyo":
		blockEngineURl = jito_go.Tokyo.BlockEngineURL
	default:
		slog.Error("BlockEngineURLLocation not valid, must be amsterdam, frankfurt, newyork or tokyo")
		return Config{}, errors.New("BlockEngineURLLocation not valid")
	}
	config.CopyTraderConfig.GlobalBuyCounter = map[string]int64{}

	// Choose the strategy to use to send transactions
	if config.CopyTraderConfig.Jito {
		slog.Info("Sending transactions using JITO")
		client, err := searcher_client.NewNoAuth(context.Background(), blockEngineURl, rpc.New(config.RPCUrl), rpc.New(config.RPCUrl), nil)
		if err != nil {
			slog.Error("Error creating searcher client: " + err.Error())
			return Config{}, err
		}
		config.CopyTraderConfig.JitoSearcherClient = client
		config.CopyTraderConfig.Sender = strategy.NewJitoSend(config.HttpClient, config.CopyTraderConfig.JitoSearcherClient, config.TPS)
	} else if config.CopyTraderConfig.Nextblock {
		slog.Info("Sending transactions using Nextblock")
		if config.CopyTraderConfig.NextblockLocation == "" {
			slog.Error("NextblockLocation not set")
			return Config{}, errors.New("NextblockLocation not set")
		}
		config.CopyTraderConfig.Sender = strategy.NewNextblockSend(config.CopyTraderConfig.NextblockLocation, config.CopyTraderConfig.NextblockAPIKEY)
	} else if config.CopyTraderConfig.Temporal {
		slog.Info("Sending transactions using Temporal")
		config.CopyTraderConfig.Sender = strategy.NewTemporalSend(config.CopyTraderConfig.TemporalURL, config.CopyTraderConfig.TemporalAPIKey)
	} else if config.CopyTraderConfig.Circular {
		slog.Info("Sending transactions using circular send")
		config.CopyTraderConfig.Sender = strategy.NewCircularSend(config.CopyTraderConfig.CircularAPIKey)
	} else if config.CopyTraderConfig.BloxRoute {
		slog.Info("Sending transactions using BloxRoute")
		config.CopyTraderConfig.Sender = strategy.NewBloxRouteSend(config.CopyTraderConfig.BloxRouteAPIKey, config.CopyTraderConfig.BloxRouteURL)
	} else {
		slog.Info("Sending transactions using normal send")
		config.CopyTraderConfig.Sender = strategy.NewNormalSend(config.TPS, config.HttpClient)
	}

	if config.CopyTraderConfig.Jito && config.CopyTraderConfig.Nextblock || config.CopyTraderConfig.Jito && config.CopyTraderConfig.Temporal || config.CopyTraderConfig.Nextblock && config.CopyTraderConfig.Temporal || config.CopyTraderConfig.Circular && config.CopyTraderConfig.Jito || config.CopyTraderConfig.Circular && config.CopyTraderConfig.Nextblock || config.CopyTraderConfig.Circular && config.CopyTraderConfig.Temporal {
		slog.Error("Cant use Jito, Nextblock and Temporal together")
		return Config{}, errors.New("cant use Jito, Nextblock and Temporal together")
	}

	if config.CopyTraderConfig.Sender == nil {
		slog.Error("Sender not set")
		return Config{}, errors.New("sender not set")
	}

	return config, nil
}

// LoadTomlFile loads the toml file and returns a Config struct
func LoadTomlFile(filename string, whitelabel string) (Config, error) {
	//Load the toml file, convert to string and call NewConfig
	file, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	conf, err := NewConfig(string(file), whitelabel)
	if err != nil {
		return Config{}, err
	}

	if verify(conf) {
		return conf, nil
	} else {
		return Config{}, errors.New("config not valid")
	}
}

func (c *Config) NewTasks(data string) error {
	var records []*CopyTraderTask
	if err := gocsv.UnmarshalString(data, &records); err != nil {
		return err
	}

	tasks := make([]CopyTraderTask, len(records))
	for i, record := range records {
		if record.DynamicFee {
			if record.DynamicFeeLevel != "min" && record.DynamicFeeLevel != "low" && record.DynamicFeeLevel != "medium" && record.DynamicFeeLevel != "high" && record.DynamicFeeLevel != "veryHigh" && record.DynamicFeeLevel != "unsafeMax" {
				return errors.New("dynamic fee level not valid")
			}
		}

		wallet, err := solana.WalletFromPrivateKeyBase58(record.PrivateKey)
		if err != nil || len(wallet.PrivateKey) <= 30 {
			return fmt.Errorf("invalid private key for task %s", record.TaskName)
		}

		tasks[i] = CopyTraderTask{
			TaskName:               record.TaskName,
			PrivateKey:             record.PrivateKey,
			WalletToCopy:           record.WalletToCopy,
			DynamicFee:             record.DynamicFee,
			DynamicFeeLevel:        record.DynamicFeeLevel,
			FixedFee:               record.FixedFee,
			SolBuyAmount:           record.SolBuyAmount,
			BuySlippage:            record.BuySlippage,
			SellingSlippage:        record.SellingSlippage,
			CopySellPercentage:     record.CopySellPercentage,
			MonitorMode:            record.MonitorMode,
			CopySell:               record.CopySell,
			FirstTimeBuyCache:      NewFirstTimeBuyCache(),
			TokensOwnedCache:       NewTokensOwnedCache(),
			Wallet:                 wallet,
			Enabled:                record.Enabled,
			TaskMaximumBuysCounter: make(map[string]int64),
			TaskMaximumBuys:        record.TaskMaximumBuys,
			MinimumMarket:          record.MinimumMarket,
			StopLoss:               record.StopLoss,
			TakeProfit:             record.TakeProfit,
			BuyPercentage:          record.BuyPercentage,
		}
	}

	c.CopyTraderConfig.AllTasks = tasks
	c.ReloadActiveTasks()
	return nil
}

func (c *Config) ReloadActiveTasks() {
	c.CopyTraderConfig.ActiveTasks = nil
	for i := 0; i < len(c.CopyTraderConfig.AllTasks); i++ {
		if c.CopyTraderConfig.AllTasks[i].Enabled {
			c.CopyTraderConfig.ActiveTasks = append(c.CopyTraderConfig.ActiveTasks, &c.CopyTraderConfig.AllTasks[i])
		}
	}
}

func (c *Config) LoadCopyTraderTasks(filename string) (err error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = c.NewTasks(string(file))
	if err != nil {
		return err
	}

	if verifyTasks(*c) {
		return nil
	} else {
		return errors.New("tasks not valid")
	}
}

func verify(conf Config) bool {
	if conf.HttpClient == nil {
		return false
	}
	if conf.CopyTraderConfig.HeliusUrl == "" {
		slog.Error("HeliusUrl not set")
		return false
	}
	if conf.CopyTraderConfig.Jito {
		if conf.CopyTraderConfig.JitoSearcherClient == nil {
			return false
		}
	} else {
		if conf.CopyTraderConfig.JitoSearcherClient != nil {
			return false
		}
	}
	if conf.CopyTraderConfig.GlobalBuyCounter == nil {
		return false
	}
	return true
}

func verifyTasks(conf Config) bool {
	if len(conf.CopyTraderConfig.AllTasks) == 0 {
		return true
	}
	if conf.CopyTraderConfig.AllTasks[0].PrivateKey == "" {
		return false
	}
	if conf.CopyTraderConfig.AllTasks[0].WalletToCopy == "" {
		return false
	}

	for i := 0; i < len(conf.CopyTraderConfig.AllTasks); i++ {
		if conf.CopyTraderConfig.AllTasks[i].Wallet == nil {
			return false
		}
		if conf.CopyTraderConfig.AllTasks[i].SellingSlippage < 0 || conf.CopyTraderConfig.AllTasks[i].SellingSlippage > 1 {
			return false
		}
		if conf.CopyTraderConfig.AllTasks[i].BuySlippage < 0 {
			return false
		}
		if conf.CopyTraderConfig.AllTasks[i].FixedFee < 0 {
			return false
		}
		if conf.CopyTraderConfig.AllTasks[i].SolBuyAmount < 0 {
			return false
		}
		if conf.CopyTraderConfig.GlobalBuyCounter == nil {
			return false
		}
		if conf.CopyTraderConfig.AllTasks[i].TaskMaximumBuys < 0 || conf.CopyTraderConfig.AllTasks[i].TaskMaximumBuys > conf.CopyTraderConfig.MaximumBuysPerToken {
			slog.Error("Task: " + conf.CopyTraderConfig.AllTasks[i].TaskName + " has invalid maximum buys")
			return false
		}
	}

	return true
}
