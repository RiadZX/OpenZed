package mocks

import (
	"github.com/gookit/slog"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test the setup to make sure it works correctly

// Test_InitTestLogger tests the InitTestLogger function.
// It initializes the logger for mocks purposes and sets up the logger to write to a custom io.Writer.
func Test_CaptureLogOutput(t *testing.T) {
	// Capture log output
	logOutput := CaptureLogOutput()

	// Log a test message
	slog.Info("Test message")
	slog.Debug("Debug message")
	slog.Error("Error message")
	slog.Warn("Warning message")

	// Assert the log output
	assert.Contains(t, logOutput.String(), "Test message")
	assert.Contains(t, logOutput.String(), "Error message")
	assert.Contains(t, logOutput.String(), "Warning message")
	assert.Contains(t, logOutput.String(), "Debug message")
}
