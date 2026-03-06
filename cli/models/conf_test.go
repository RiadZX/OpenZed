package models

import (
	"testing"

	"github.com/gagliardetto/solana-go/rpc"
)

func TestLoadingConf(t *testing.T) {
	var data = `
rpcUrl="https://nothelius/?api-key=XYZ"
wssUrl="wss://mainnet.helius-rpc.com/?api-key=XYZ"
blockEngineURLLocation="amsterdam"
TPS=500
pnlDelayMS=1000
[copyTrader]
jito=true
tipFee=0.1
heliusSending=false
heliusUrl="hhttps://mainnet.helius-rpc.com/?api-key=XYZ"
getTransactionDelayMS = 50
getTransactionRetries = 75
getAccountInfoDelayMS = 50
getAccountInfoRetries = 75
`
	conf, err := NewConfig(data, "Zed")
	if err != nil {
		t.Fatal(err)
	}

	if conf.RPCUrl != "https://nothelius/?api-key=XYZ" {
		t.Fatal("RPCUrl not loaded correctly")
	}
	if conf.WSSUrl != "wss://mainnet.helius-rpc.com/?api-key=XYZ" {
		t.Fatal("WSSUrl not loaded correctly")
	}

	if conf.CopyTraderConfig.GetAccountInfoDelayMS != 50 {
		t.Fatal("GetAccountInfoDelayMS not loaded correctly, got:", conf.CopyTraderConfig.GetAccountInfoDelayMS)
	}

	if conf.CopyTraderConfig.GetAccountInfoRetries != 75 {
		t.Fatal("GetAccountInfoRetries not loaded correctly, got:", conf.CopyTraderConfig.GetAccountInfoRetries)
	}
	//check if client and wallet are initialized
	if conf.HttpClient == nil {
		t.Fatal("HttpClient not initialized")
	}
	if conf.TPS != 500 {
		t.Fatal("TPS not loaded correctly")
	}

	if verify(conf) == false {
		t.Fatal("Config not valid")
	}

	if verifyTasks(conf) == false {
		t.Fatal("AllTasks not valid")
	}
}

func TestLoadingTasks(t *testing.T) {
	var tasks = `taskname,privatekey,wallet_to_copy,dynamic_fee,dynamic_fee_level,fixed_fee,sol_buy_amount,buy_slippage,selling_slippage,task_maximum_buys_per_token,copy_sell_percentage,monitormode,enabled
test1,<PRIVATEKEY>,4fQ7ucQyA6E5Y2y7Jw3tjx4R1Qk9U5kjLy4EJGNLsVdQ,true,low,0.005,0.0001,0.05,0.05,5,true,true,true
test2,<PRIVATEKEY>,4fQ7ucQyA6E5Y2y7Jw3tjx4R1Qk9U5kjLy4EJGNLsVdQ,true,unsafeMax,0.005,0.0001,0.05,0.05,1,false,true,true`
	conf := Config{
		CopyTraderConfig: &CopyTraderConfig{},
	}

	err := conf.NewTasks(tasks)
	if err != nil {
		t.Fatal(err)
	}

	if len(conf.CopyTraderConfig.AllTasks) != 2 {
		t.Fatal("AllTasks not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].TaskName != "test1" {
		t.Fatal("TaskName not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].DynamicFeeLevel != "low" {
		t.Fatal("DynamicFeeLevel not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].FixedFee != 0.005 {
		t.Fatal("FixedFee not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].SolBuyAmount != 0.0001 {
		t.Fatal("SolBuyAmount not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].MonitorMode != true {
		t.Fatal("SolBuyAmount not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].Enabled != true {
		t.Fatal("Task not enabled")
	}

	if conf.CopyTraderConfig.AllTasks[0].CopySellPercentage != true {
		t.Fatal("Task not enabled")
	}
	if conf.CopyTraderConfig.AllTasks[1].CopySellPercentage != false {
		t.Fatal("Task not enabled")
	}

	if conf.CopyTraderConfig.ActiveTasks[0].Enabled != true {
		t.Fatal("Task not enabled")
	}

	if conf.CopyTraderConfig.ActiveTasks[1].Enabled != true {
		t.Fatal("Task not enabled")
	}
	if conf.CopyTraderConfig.ActiveTasks[1].TaskMaximumBuys != 1 {
		t.Fatal("TaskMaximumBuys not loaded correctly")
	}
	if conf.CopyTraderConfig.ActiveTasks[0].TaskMaximumBuys != 5 {
		t.Fatal("TaskMaximumBuys not loaded correctly")
	}

}

func TestConfigAndTasks(t *testing.T) {
	var data = `
rpcUrl="https://nothelius/?api-key=XYZ"
wssUrl="wss://mainnet.helius-rpc.com/?api-key=XYZ"
blockEngineURLLocation="newyork"
pnlDelayMS=1000
[copyTrader]
global_maximum_buys_per_token=5
jito=true
heliusSending=true
heliusUrl="hhttps://mainnet.helius-rpc.com/?api-key=XYZ"
getTransactionDelayMS = 50
getTransactionRetries = 75
getAccountInfoDelayMS = 50
getAccountInfoRetries = 75
`
	var tasks = `taskname,privatekey,wallet_to_copy,dynamic_fee,dynamic_fee_level,fixed_fee,sol_buy_amount,buy_slippage,selling_slippage,task_maximum_buys_per_token,copy_sell_percentage,monitormode,enabled
test1,<PRIVATEKEY>,4fQ7ucQyA6E5Y2y7Jw3tjx4R1Qk9U5kjLy4EJGNLsVdQ,true,low,0.005,0.0001,0.05,0.05,3,false,true,true
test2,<PRIVATEKEY>,4fQ7ucQyA6E5Y2y7Jw3tjx4R1Qk9U5kjLy4EJGNLsVdQ,true,unsafeMax,0.005,0.0001,0.05,0.05,2,true,true,false`

	conf, err := NewConfig(data, "Zed")
	if err != nil {
		t.Fatal(err)
	}

	err = conf.NewTasks(tasks)
	if err != nil {
		t.Fatal(err)
	}

	if verify(conf) == false {
		t.Fatal("Config not valid")
	}

	if verifyTasks(conf) == false {
		t.Fatal("AllTasks not valid")
	}

	if conf.CopyTraderConfig.AllTasks[0].TaskName != "test1" {
		t.Fatal("TaskName not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].DynamicFeeLevel != "low" {
		t.Fatal("DynamicFeeLevel not loaded correctly")
	}

	if conf.WSSUrl != "wss://mainnet.helius-rpc.com/?api-key=XYZ" {
		t.Fatal("WSSUrl not loaded correctly")
	}

	if conf.CopyTraderConfig.GetAccountInfoDelayMS != 50 {
		t.Fatal("GetAccountInfoDelayMS not loaded correctly, got:", conf.CopyTraderConfig.GetAccountInfoDelayMS)
	}

	if conf.CopyTraderConfig.GetAccountInfoRetries != 75 {
		t.Fatal("GetAccountInfoRetries not loaded correctly, got:", conf.CopyTraderConfig.GetAccountInfoRetries)
	}
	//check if client and wallet are initialized
	if conf.HttpClient == nil {
		t.Fatal("HttpClient not initialized")
	}

	if conf.CopyTraderConfig.AllTasks[0].Enabled != true {
		t.Fatal("Task not enabled")
	}
	if conf.CopyTraderConfig.AllTasks[1].Enabled != false {
		t.Fatal("Task not disabled")
	}

	//check active tasks
	if len(conf.CopyTraderConfig.ActiveTasks) != 1 {
		t.Fatal("ActiveTasks not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].TaskMaximumBuys != 3 {
		t.Fatal("TaskMaximumBuys not loaded correctly")
	}
}

func TestTasksToCSVString(t *testing.T) {
	var tasks = []CopyTraderTask{
		{
			TaskName:           "test1",
			PrivateKey:         "<PRIVATEKEY>",
			WalletToCopy:       "b",
			DynamicFee:         true,
			DynamicFeeLevel:    "low",
			FixedFee:           0.005,
			SolBuyAmount:       0.0001,
			BuySlippage:        0.05,
			SellingSlippage:    0.05,
			MonitorMode:        true,
			Enabled:            true,
			CopySellPercentage: true,
			TaskMaximumBuys:    3,
		}, {
			TaskName:           "test5",
			PrivateKey:         "<PRIVATEKEY>",
			WalletToCopy:       "b",
			DynamicFee:         false,
			DynamicFeeLevel:    "low",
			FixedFee:           0.005,
			SolBuyAmount:       0.0001,
			BuySlippage:        0.05,
			SellingSlippage:    0.05,
			MonitorMode:        true,
			Enabled:            true,
			CopySellPercentage: false,
			TaskMaximumBuys:    0,
		},
	}

	csvString := tasksToCSVString(tasks)
	expected := "taskname,privatekey,wallet_to_copy,dynamic_fee,dynamic_fee_level,fixed_fee,sol_buy_amount,buy_slippage,selling_slippage,task_maximum_buys_per_token,copy_sell_percentage,monitormode,enabled\n" +
		"test1,<PRIVATEKEY>,b,true,low,0.005000,0.000100,0.050000,0.050000,3,true,true,true\n" +
		"test5,<PRIVATEKEY>,b,false,low,0.005000,0.000100,0.050000,0.050000,0,false,true,true\n"
	if csvString != expected {
		t.Fatal("CSV string not correct, got:", csvString)
	}

	//from string to tasks
	conf := Config{
		CopyTraderConfig: &CopyTraderConfig{},
	}

	err := conf.NewTasks(csvString)
	if err != nil {
		t.Fatal(err)
	}

	if len(conf.CopyTraderConfig.AllTasks) != 2 {
		t.Fatal("AllTasks not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].FixedFee != 0.005 {
		t.Fatal("FixedFee not loaded correctly")
	}

	if conf.CopyTraderConfig.AllTasks[0].SolBuyAmount != 0.0001 {
		t.Fatal("Tasks not loaded correctly")
	}

}

func Benchmark_createconnections(b *testing.B) {
	var url = "https://mainnet.helius-rpc.com/"
	for i := 0; i < b.N; i++ {
		_ = rpc.New(url)
	}
}
