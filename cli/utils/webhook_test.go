package utils

import "Zed/ui"

func init() {
	ui.WhiteLabel = "Zed"
	ui.InitBranding()
}

//
//func Test_SendWebhook(t *testing.T) {
//	webhook := "https://discord.com/api/webhooks/WEBHOOKURL"
//	swap := models.SwapInfo{
//		License:              "TEST",
//		Timestamp:            "213312312",
//		UserWallet:           "dawdawdwa",
//		ProgramName:          "PumpFun",
//		IsBuy:                true,
//		TokenMint:            "123123",
//		TokenMintAmount:      12312.45,
//		SolanaAmount:         0.12,
//		TransactionSignature: "dddd",
//	}
//	SendAnonWebhook(swap)
//	SendWebhook(webhook, swap, false)
//}
