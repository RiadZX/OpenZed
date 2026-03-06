package ui

import (
	"Zed/db"
	"Zed/models"
	"fmt"
	"strings"
	"time"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/manifoldco/promptui"
)

var logo = cyan(` 
 ▄███████▄     ▄████████ ████████▄  
 ██▀     ▄██   ███    ███ ███   ▀███
       ▄███▀   ███    █▀  ███    ███
  ▀█▀▄███▀▄▄  ▄███▄▄▄     ███    ███
   ▄███▀   ▀ ▀▀███▀▀▀     ███    ███
 ▄███▀         ███    █▄  ███    ███
 ███▄     ▄█   ███    ███ ███   ▄███
  ▀████████▀   ██████████ ████████▀
`)

type ZedUI struct {
	dbConnection *db.Connection
}

// Returns initialized ZedUI struct
func NewZedUI() *ZedUI {
	var z = ZedUI{}
	z.InitLogger()
	err := z.initDBConnection()
	if err != nil {
		slog.Error(err)
		return nil
	}
	return &z
}

func (z *ZedUI) GetLogo() string {
	return logo
}

func (z *ZedUI) GetWebhookName() string {
	return "Zed Software"
}

func (z *ZedUI) MainMenu() string {
	items := []string{
		cyan("PumpFun+Raydium Copytrader"),
		cyan("PumpFun Copytrader"),
		cyan("Raydium Copytrader"),
	}

	items = append(items, cyan("Exit"))
	prompt := promptui.Select{
		Label: blue("Select module"),
		Items: items,
		Size:  len(items),
	}
	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "INVALID"
	}

	return items[i]
}

func (z *ZedUI) LogConfigData(userConf *models.Config) {
	var grpc bool
	if userConf.GRPCUrl != "" {
		grpc = true
	}
	slog.Info("Active Tasks: " + fmt.Sprint(len(userConf.CopyTraderConfig.ActiveTasks)))
	for i := 0; i < len(userConf.CopyTraderConfig.ActiveTasks); i++ {
		slog.Info(cyan("Task "+fmt.Sprint(i)+": ") + userConf.CopyTraderConfig.ActiveTasks[i].TaskName)
		// sp/tp
		slog.Info(yellow("Stoploss: ") + fmt.Sprint(userConf.CopyTraderConfig.ActiveTasks[i].StopLoss))
		slog.Info(yellow("Takeprofit: ") + fmt.Sprint(userConf.CopyTraderConfig.ActiveTasks[i].TakeProfit))
	}
	if grpc {
		slog.Info(cyan("Using GRPC"))
		slog.Info(cyan("GRPC URL: ") + userConf.GRPCUrl)
	} else {
		slog.Info("Using Websockets")
		slog.Info(cyan("RPC URL: ") + userConf.RPCUrl)
		slog.Info(cyan("Websocket URL: ") + userConf.WSSUrl)
		for _, r := range userConf.ExtraRPCS {
			slog.Info(cyan("Extra RPC URL: ") + r)
		}
	}
}

func (z *ZedUI) InitLogger() {
	fileName := "app" + time.Now().Format("2006-01-02 15:04:05") + ".log"
	//replace : -
	fileName = strings.Replace(fileName, ":", "_", -1)
	fileName = strings.Replace(fileName, " ", "_", -1)
	fileName = strings.Replace(fileName, "-", "_", -1)
	fileHandler, err := handler.NewFileHandler(fileName, handler.WithLogLevels(slog.AllLevels))
	if err != nil {
		panic(err)
	}
	logFormat := "{{datetime}} [{{level}}] {{message}}\n"
	fileLogFormat := "{{datetime}} [{{caller}}] [{{level}}] {{message}}\n"
	fileHandler.SetFormatter(slog.NewTextFormatter(fileLogFormat))
	slog.SetFormatter(slog.NewTextFormatter(logFormat))
	//enable colors
	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
	})
	slog.SetLogLevel(slog.InfoLevel)
	slog.PushHandlers(fileHandler)
}

func (z *ZedUI) initDBConnection() error {
	retries := 5
	for i := 0; i < retries; i++ {
		con, err := db.NewDBConnection("Zed")
		if err != nil {
			return err
		}
		z.dbConnection = &con
	}

	return nil
}

func (z *ZedUI) GetDBConnection() *db.Connection {
	if z.dbConnection == nil {
		err := z.initDBConnection()
		if err != nil {
			slog.Error(err)
			return nil
		}
	}
	return z.dbConnection
}

func (z *ZedUI) Authenticate(licenseKey string) bool {
	// AUTHENTICATION DISABLED - As the servers are down.
	return true

	if licenseKey == "" {
		slog.Error("License key not provided, please provide a valid license key")
		return false
	}
	slog.Info("Authenticating...")

	var license *db.License
	err := retry(5, 1*time.Second, func() error {
		var err error
		license, err = z.GetDBConnection().GetLicenseData(licenseKey)
		if err != nil && err.Error() == "license not found" {
			slog.Error("License not found, please provide a valid license key")
		}
		return err
	})
	if err != nil {
		slog.Error(err)
		return false
	}

	//check if license is expired
	if license.EndDate < time.Now().Unix() {
		slog.Error("License expired")
		return false
	}
	if license.Status != "active" && license.Status != "trial" {
		slog.Error("License not active")
		return false
	}

	if license.IP != "" {
		//hardware id registered, check if it matches
		ip, err := GetOutboundIP()
		if err != nil {
			slog.Error(err)
			return false
		}
		if license.IP != ip {
			slog.Error("IP mismatch, please reset IP")
			return false
		}
	} else {
		slog.Info("IP not registered, registering...")
		//hardware id not registered, register it
		ip, err := GetOutboundIP()
		if err != nil {
			slog.Error(err)
			return false
		}
		//register it
		err = z.GetDBConnection().RegisterIP(license, ip)
		if err != nil {
			slog.Error(err)
			return false
		}
		slog.Info("IP registered successfully!")
	}

	userData, err := GetUserData(license.DiscordID)
	if err != nil {
		userData = &UserDataResponse{Username: "User"}
	}

	slog.Info("License: " + license.Status)
	slog.Info("License active until: " + time.Unix(license.EndDate, 0).Format("2006-01-02 15:04:05"))
	slog.Info("Authentication successful, Welcome " + userData.Username + "!")
	return true
}
