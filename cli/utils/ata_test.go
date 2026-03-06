package utils

import "testing"

func Test_utils_getATA(t *testing.T) {
	//Data needed:
	ata, _ := GetATA("ZDK6SbnjEtgTWG1CqHhNueEFvziCcMQYQhW2UwJ3EzU", "2J8L9hdrGE8Lt5EURhg62gNLymGf7onibFJX2MVxpump")
	if ata != "FaMQghoHD9EtyyLnX4JKB5c7ZxE3PxfHsjDb1kPGRehh" {
		t.Fatal("ATA not found")
	}
}
