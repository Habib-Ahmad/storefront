package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"golang.org/x/term"
)

// New returns a structured slog.Logger.
// Production: JSON to stdout. Development: human-readable text to stdout.
// level is read from the LOG_LEVEL env var (debug|info|warn|error), defaulting to info.
func New(env, level string) *slog.Logger {
	handlerLevel := parseLevel(level)

	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: handlerLevel})
	} else {
		h = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      handlerLevel,
			TimeFormat: "15:04:05",
			NoColor:    disableColor(),
		})
	}

	return slog.New(h)
}

func disableColor() bool {
	return !term.IsTerminal(int(os.Stdout.Fd()))
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
