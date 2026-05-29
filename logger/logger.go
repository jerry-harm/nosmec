package logger

import (
	"io"
	"log/slog"
	"os"

	"fiatjaf.com/nostr"
)

var (
	logger    *slog.Logger
	debugFile *os.File
)

func init() {
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger = slog.New(handler)

	nostr.InfoLogger.SetOutput(io.Discard)
	nostr.DebugLogger.SetOutput(io.Discard)
}

func SetDebug(enabled bool) {
	if enabled {
		f, err := os.OpenFile("nosmec-debug.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			return
		}
		debugFile = f
		handler := slog.NewJSONHandler(f, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger = slog.New(handler)
		nostr.InfoLogger.SetOutput(f)
		nostr.DebugLogger.SetOutput(f)
	} else {
		if debugFile != nil {
			debugFile.Close()
			debugFile = nil
		}
		nostr.InfoLogger.SetOutput(io.Discard)
		nostr.DebugLogger.SetOutput(io.Discard)
		handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		})
		logger = slog.New(handler)
	}
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	logger.Error(msg, args...)
	os.Exit(1)
}