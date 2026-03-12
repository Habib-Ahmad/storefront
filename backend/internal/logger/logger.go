package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a structured slog.Logger.
// Production: JSON to stdout. Development: human-readable text to stdout.
// level is read from the LOG_LEVEL env var (debug|info|warn|error), defaulting to info.
func New(env, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(h)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
