package utils

import (
	"testing"
)

func TestGetTimeString(t *testing.T) {
	timestamp := GetTimeString()
	if timestamp == "" {
		t.Error("Expected a non-empty string")
	}
}
