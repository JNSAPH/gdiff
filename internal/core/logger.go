package core

import (
	"log/slog"
	"os"
)

// LogFilePath is where gdiff writes its logs in dev mode. The TUI owns the
// terminal (alt screen), so logs can't go to stderr while it runs.
const LogFilePath = "gdiff.log"

// SetupLogger writes slog to LogFilePath in dev mode and discards it
// otherwise — a normal run must not litter the repo it was pointed at.
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
