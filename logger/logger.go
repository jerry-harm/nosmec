package logger

import (
	"io"
	"log/slog"
	"os"

	"fiatjaf.com/nostr"
)

var (
	logger  *slog.Logger
	verbose bool
)

func init() {
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger = slog.New(handler)

	nostr.InfoLogger.SetOutput(io.Discard)
	nostr.DebugLogger.SetOutput(io.Discard)
}

func SetVerbose(v bool) {
	verbose = v
	if v {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		nostr.InfoLogger.SetOutput(os.Stderr)
		nostr.DebugLogger.SetOutput(os.Stderr)
	} else {
		nostr.InfoLogger.SetOutput(io.Discard)
		nostr.DebugLogger.SetOutput(io.Discard)
	}
}

func IsVerbose() bool {
	return verbose
}

func Debug(msg string, args ...any) {
	if verbose {
		logger.Debug(msg, args...)
	}
}

func Info(msg string, args ...any) {
	if verbose {
		logger.Info(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if verbose {
		logger.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if verbose {
		logger.Error(msg, args...)
	}
}

func Fatal(msg string, args ...any) {
	logger.Error(msg, args...)
	os.Exit(1)
}
