package logger

import (
	"log/slog"
	"os"
)

// New returns a structured slog.Logger.
// Production: JSON to stdout. Development: human-readable text to stdout.
func New(env string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}

	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(h)
}
