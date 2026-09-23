package logger

import (
	"log/slog"
	"os"
)

func Setup(environment, logLevel string) {
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if environment == "production" {
		// JSON output for production — structured logs for tools like
		// Datadog, CloudWatch, or ELK stack.
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text output for development — easier to read in the terminal.
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
