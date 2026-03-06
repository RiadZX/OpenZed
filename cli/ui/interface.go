package ui

import (
	"Zed/db"
	"Zed/models"
)

type UI interface {
	// GetLogo returns the ASCII art logo for the application.
	GetLogo() string
	// GetWebhookName returns the name of the webhook for the application.
	GetWebhookName() string
	// MainMenu displays the main menu for the application and returns the selected module.
	// The module is selected using a promptui.Select prompt.
	MainMenu() string
	LogConfigData(userConf *models.Config)
	// InitLogger initializes the logger for the application.
	// It sets up a file handler to log all levels to a file with a timestamped filename.
	// **Debugs are only logged in the file and not to the terminal**
	// The log format for the console and file are set, and color is enabled for console logs.
	InitLogger()
	GetDBConnection() *db.Connection
	//Authenticate checks the license key against the server and returns true if the license is valid.
	Authenticate(licenseKey string) bool
}
