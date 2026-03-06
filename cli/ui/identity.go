package ui

//Declare enum, this will be used to keep track of which whitelabel is being used.
//whitelabel choosing will be passed as a flag to the program on compile time. the program will behave differently based on the whitelabel chosen.
//this file will act as a store of all whitelabel related flags that will be passed

var WhiteLabel string
var AnonWebhookURL string
var AvatarURL string
var ZedAuthorURL string

func InitBranding() {
	if WhiteLabel != "Zed" {
		panic("Invalid Brand")
	}
	ZedAuthorURL = "https://pbs.twimg.com/profile_images/1824399529031929856/Vb-q0XXn_400x400.jpg"
	if WhiteLabel == "Zed" {
		AnonWebhookURL = "https://discord.com/api/webhooks/WEBHOOKURL"
		AvatarURL = "https://pbs.twimg.com/profile_images/1824399529031929856/Vb-q0XXn_400x400.jpg"
	}
}

func GetWebhookName() string {
	return WhiteLabel + " Software"
}
