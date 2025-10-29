package logging

import (
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"
)

func NewLogger() *slog.Logger {
	options := slog.HandlerOptions{Level: getLogLevel()}

	var handler slog.Handler
	format := os.Getenv("LDPA_LOG_FORMAT")

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &options)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, &options)
	default:
		if term.IsTerminal(int(os.Stdout.Fd())) {
			handler = slog.NewTextHandler(os.Stdout, &options)
		} else {
			handler = slog.NewJSONHandler(os.Stdout, &options)
		}
	}

	return slog.New(handler)
}

func getLogLevel() slog.Level {
	level := os.Getenv("LDPA_LOG_LEVEL")

	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
