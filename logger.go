package main

import (
	"log/slog"
	"os"
)

// newAppLogger creates a new logger for the app
func newAppLogger(verbose bool) *slog.Logger {

	logHandlerOptions := &slog.HandlerOptions{
		Level: slog.LevelError,
	}

	if verbose {
		logHandlerOptions.Level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, logHandlerOptions))

	return logger
}
