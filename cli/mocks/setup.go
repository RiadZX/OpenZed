package mocks

import (
	"bytes"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"io"
)

// InitTestLogger initializes the logger for mocks purposes.
// It sets up the logger to write to a custom io.Writer (e.g., bytes.Buffer).
func InitTestLogger(writer io.Writer) {
	// Create a new handler that writes to the provided writer
	testHandler := handler.NewIOWriterHandler(writer, []slog.Level{slog.InfoLevel, slog.ErrorLevel, slog.WarnLevel, slog.DebugLevel})

	// Define log format for the test handler
	logFormat := "{{datetime}} [{{level}}] {{message}}\n"
	testHandler.SetFormatter(slog.NewTextFormatter(logFormat))

	// Clear existing handlers and set the test handler
	slog.Reset()
	slog.PushHandlers(testHandler)

	// Set the log level to Info
	slog.SetLogLevel(slog.InfoLevel)
}

// CaptureLogOutput captures the log output for assertions in tests.
// It returns a bytes.Buffer that contains the log output.
func CaptureLogOutput() *bytes.Buffer {
	var buf bytes.Buffer
	InitTestLogger(&buf)
	return &buf
}
