package core

import (
	"log/slog"
	"os"
)

// LogFilePath is where dev-mode logs go: the TUI owns the terminal, not stderr.
const LogFilePath = "gdiff.log"

// SetupLogger logs to LogFilePath in dev mode and discards otherwise — a normal
// run must not litter the repo it was pointed at.
func SetupLogger(dev bool) {
	if !dev {
		slog.SetDefault(slog.New(slog.DiscardHandler))
		return
	}

	file, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Warn("could not open log file, logging to stderr instead", "error", err)
		return
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug})))
}
