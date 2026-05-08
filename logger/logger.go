package logger

import (
	"log/slog"
	"os"
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
}

func SetVerbose(v bool) {
	verbose = v
	if v {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
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
	logger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	logger.Error(msg, args...)
	os.Exit(1)
}
