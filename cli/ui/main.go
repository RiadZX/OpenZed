package ui

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/gookit/slog"
	"math"
	"net/http"
	"time"
)

var (
	red     = color.FgRed.Render
	green   = color.FgGreen.Render
	yellow  = color.FgYellow.Render
	blue    = color.FgBlue.Render
	magenta = color.FgMagenta.Render
	cyan    = color.FgCyan.Render
	white   = color.FgWhite.Render
	bold    = color.OpBold.Render
)

// GetOutboundIP Get preferred outbound ip of this machine
func GetOutboundIP() (string, error) {
	// Get public ip using ipify
	//supports both ipv4 and ipv6
	resp, err := http.Get("https://api64.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip := ""
	_, err = fmt.Fscan(resp.Body, &ip)
	if err != nil {
		return "", err
	}
	return ip, nil
}

// Exponential backoff retry function
func retry(attempts int, sleep time.Duration, fn func() error) error {
	for i := 0; ; i++ {
		err := fn()
		if err == nil || err.Error() == "license not found" {
			return nil
		}
		if i >= (attempts - 1) {
			return err
		}
		slog.Warnf("Attempt %d failed; retrying in %v", i+1, sleep)
		time.Sleep(sleep)
		sleep = time.Duration(float64(sleep) * math.Pow(2, float64(i)))
	}
}
