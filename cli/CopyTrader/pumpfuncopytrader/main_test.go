package pumpfuncopytrader

import (
	"Zed/mocks"
	"Zed/models"
	"Zed/pumpfunsdk"
	"Zed/utils"
	"errors"
	"fmt"
	"testing"
)

func Test_getPurchaseDataCustomProgram(t *testing.T) {
	tradeData := pumpfunsdk.TradeEventData{IsBuy: true}
	tx, inner := mocks.GetPFSwapCustomProgramBuyTransaction()
	purchaseData, err := getPurchaseData(&tradeData, &tx, &inner)
	if err != nil {
		t.Fatal(err)
	}
	if "rE3JG5gUXzGUvqSX2CvydsAzm1bZ4FsFgXocpyYXRA6" != purchaseData.BondingCurve {
		t.Fail()
	}
	if "HV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4" != purchaseData.AssociatedBondingCurve {
		t.Fail()
	}
}

func Test_getPurchaseData(t *testing.T) {
	tradeData := pumpfunsdk.TradeEventData{IsBuy: true}
	tx, inner := mocks.GetPFSwapBuyTransaction()
	purchaseData, err := getPurchaseData(&tradeData, &tx, &inner)
	if err != nil {
		t.Fatal(err)
	}
	if "rE3JG5gUXzGUvqSX2CvydsAzm1bZ4FsFgXocpyYXRA6" != purchaseData.BondingCurve {
		t.Fail()
	}
	if "HV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4" != purchaseData.AssociatedBondingCurve {
		t.Fail()
	}
}

func Test_shouldProcessCustomProgramSell(t *testing.T) {
	tradeData := pumpfunsdk.TradeEventData{IsBuy: false}
	tx, inner := mocks.GetPFSwapCustomProgramSellTransaction()
	purchaseData, err := getPurchaseData(&tradeData, &tx, &inner)
	if err != nil {
		t.Fatal(err)
	}
	if "rE3JG5gUXzGUvqSX2CvydsAzm1bZ4FsFgXocpyYXRA6" != purchaseData.BondingCurve {
		t.Fail()
	}
	if "HV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4" != purchaseData.AssociatedBondingCurve {
		t.Fail()
	}
	//isbuy=true should fail
	tradeData = pumpfunsdk.TradeEventData{IsBuy: true}
	_, err = getPurchaseData(&tradeData, &tx, &inner)
	if err == nil {
		t.Fail()
	}
}

func Test_getata(t *testing.T) {
	publickkey := "ZDK6SbnjEtgTWG1CqHhNueEFvziCcMQYQhW2UwJ3EzU"
	mint := "2UayC2unmmaShSfG38BAd84BeiazRJSDtSPNpdkQ4A1E"

	ata, _ := utils.GetATA(publickkey, mint)
	if ata != "CDkB44GPRTAK8GM5bviCmcnToCpcbyKRXo7sZ7dPmTqo" {
		fmt.Println(ata)
		t.Fail()
	}
}

func Test_getata2(t *testing.T) {
	publickkey := "CMf6h6wya3AZKLTsMYAH2ovfKdJGbxUfTwNmWRrPxgdY"
	mint := "So11111111111111111111111111111111111111112"

	ata, _ := utils.GetATA(publickkey, mint)
	if ata != "CNs5L5iubGFpA6Zc9SHN9dnsKRnsupxtf3guJxsBwHDS" {
		fmt.Println(ata)
		t.Fail()
	}
}

func TestPFCopyTrader_Init_Invalid(t *testing.T) {
	copyTrader := &PFCopyTrader{}
	err := copyTrader.Init()
	if !errors.Is(err, ErrConfigNotAvailable) {
		t.Errorf("Expected error: %v, got: %v", ErrConfigNotAvailable, err)
	}
	if copyTrader.initialized == true {
		t.Errorf("Expected initialized to be false, got: %v", copyTrader.initialized)
	}
}

// Tests if the Init succeeds when the PFCopyTrader is correctly initialized
func TestPFCopyTrader_Init_Valid(t *testing.T) {
	copyTrader := &PFCopyTrader{
		UserConfig: &models.Config{
			RPCUrl: "https://solana.mainnet.com",
			WSSUrl: "wss://solana.mainnet.com",
		},
		CopyTraderConfig: &models.CopyTraderConfig{
			ActiveTasks: make([]*models.CopyTraderTask, 0),
		},
	}
	err := copyTrader.Init()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if copyTrader.initialized != true {
		t.Errorf("Expected initialized to be true, got: %v", copyTrader.initialized)
	}
}

// Tests if the Start fails when the PFCopyTrader is not correctly initialized
func TestPFCopyTrader_Start_NotInitialized(t *testing.T) {
	copyTrader := &PFCopyTrader{UserConfig: &models.Config{}}
	err := copyTrader.Start()
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("Expected error: %v, got: %v", ErrNotInitialized, err)
	}
}
