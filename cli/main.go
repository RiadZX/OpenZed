package main

import (
	copytrader "Zed/CopyTrader"
	"Zed/CopyTrader/pumpfun_amm_copytrader"
	"Zed/CopyTrader/pumpfuncopytrader"
	"Zed/CopyTrader/raydiumcopytrader"
	"Zed/models"
	"Zed/ui"
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/gookit/slog"
)

func pressEnter() {
	fmt.Println("Press Enter to continue...")
	_, err := fmt.Scanln()
	if err != nil {
		return
	}
}

func removeANSI(input string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(input, "")
}

func main() {
	ui.InitBranding()
	//Load the toml file
	userConfig, err := models.LoadTomlFile(filepath.Join("config.toml"), ui.WhiteLabel)
	if err != nil || reflect.DeepEqual(userConfig, models.Config{}) {
		slog.Error("Error loading config.toml: " + err.Error())
		pressEnter()
		return
	}
	var app ui.UI

	switch ui.WhiteLabel {
	case "Zed":
		app = ui.NewZedUI()
	default:
		slog.Error("Invalid WhiteLabel")
		pressEnter()
	}

	if app == nil {
		slog.Error("Error initializing UI")
		pressEnter()
		return
	}
	fmt.Println(app.GetLogo())

	if !app.Authenticate(userConfig.LicenseKey) {
		slog.Error("Authentication failed")
		pressEnter()
		return
	}

	pf, err := pumpfuncopytrader.NewPFCopyTrader(&userConfig, userConfig.CopyTraderConfig)
	if err != nil {
		slog.Error(err.Error())
		pressEnter()
		return
	}

	rdym, err := raydiumcopytrader.NewRaydiumCopyTrader(&userConfig, userConfig.CopyTraderConfig)
	if err != nil {
		slog.Error(err.Error())
		pressEnter()
		return
	}

	pf_amm, err := pumpfun_amm_copytrader.NewPfAmmCt(&userConfig, userConfig.CopyTraderConfig)
	if err != nil {
		slog.Error(err.Error())
		pressEnter()
		return
	}

	var grpc bool
	if userConfig.GRPCUrl != "" {
		grpc = true
	}

	//Load the tasks
	err = userConfig.LoadCopyTraderTasks(filepath.Join("tasks.csv"))
	if err != nil {
		slog.Error("Error loading tasks.csv: " + err.Error())
		pressEnter()
		return
	}

	app.LogConfigData(&userConfig)
	for {
		x := app.MainMenu()
		cleanedChoice := removeANSI(x)
		if x == "INVALID" {
			fmt.Println("Invalid choice")
			continue
		}
		switch cleanedChoice {
		case "PumpFun+Raydium Copytrader":
			slog.Info("Starting Raydium+PumpFun CopyTrader...")

			err = copytrader.LaunchCopyTrader(context.Background(), &userConfig, []copytrader.CopyTrader{&rdym, &pf, &pf_amm}, grpc, app.GetDBConnection())
			if err != nil {
				slog.Error(err.Error())
				continue
			}
		case "PumpFun Copytrader":
			slog.Info("Starting PumpFun CopyTrader...")
			err = copytrader.LaunchCopyTrader(context.Background(), &userConfig, []copytrader.CopyTrader{&pf, &pf_amm}, grpc, app.GetDBConnection())
			if err != nil {
				slog.Error(err.Error())
				continue
			}
		case "Raydium Copytrader":
			slog.Info("Starting Raydium CopyTrader...")
			err = copytrader.LaunchCopyTrader(context.Background(), &userConfig, []copytrader.CopyTrader{&rdym}, grpc, app.GetDBConnection())
			if err != nil {
				slog.Error(err.Error())
				continue
			}
		case "Exit":
			slog.Info("Exiting...")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}
