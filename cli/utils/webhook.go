package utils

import (
	"Zed/models"
	"Zed/ui"
	"encoding/json"
	discordwebhook "github.com/bensch777/discord-webhook-golang"
	"github.com/gookit/slog"
	"time"
)

// SendWebhook sends a webhook message to the specified URL
// It is always branded around ZED Software
func SendWebhook(url string, info models.SwapInfo, isAnon bool) {
	color := 0x00ff00
	if !info.IsBuy {
		color = 0xff0000
	}

	title := ui.GetWebhookName()
	authorUrl := ui.AvatarURL
	authorName := ui.WhiteLabel
	if isAnon {
		title = "Zed Software"
		authorUrl = ui.ZedAuthorURL
		authorName = "Zed"
	}
	description := "A buy has been made!"
	if !info.IsBuy {
		description = "A sell has been made!"
	}

	embed := discordwebhook.Embed{
		Title: title,
		Author: discordwebhook.Author{
			Name:     authorName,
			Icon_URL: authorUrl,
		},
		Color:     color,
		Timestamp: time.Now(),
		//If its a buy, then the description is "A buy has been made!",else its a sell
		Description: description,
	}

	// General fields
	embed.Fields = append(embed.Fields, discordwebhook.Field{
		Name:   "Program",
		Value:  info.ProgramName,
		Inline: false,
	})
	embed.Fields = append(embed.Fields, discordwebhook.Field{
		Name:   "Token Mint",
		Value:  info.TokenMint,
		Inline: false,
	})
	// Add fields that are for non-anon only
	if !isAnon {
		embed.Fields = append(embed.Fields, discordwebhook.Field{
			Name:   "User Wallet",
			Value:  info.UserWallet,
			Inline: false,
		})
		embed.Fields = append(embed.Fields, discordwebhook.Field{
			Name:   "Transaction Signature",
			Value:  info.TransactionSignature,
			Inline: false,
		})
	}

	hook := discordwebhook.Hook{
		Username:   title,
		Avatar_url: authorUrl,
		Embeds:     []discordwebhook.Embed{embed},
	}

	payload, err := json.Marshal(hook)
	if err != nil {
		slog.Error(err.Error())
	}
	err = discordwebhook.ExecuteWebhook(url, payload)
	if err != nil {
		slog.Error(err.Error())
	}
}

func SendAnonWebhook(info models.SwapInfo) {
	SendWebhook(ui.AnonWebhookURL, info, true)
}
