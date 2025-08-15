package logger

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger() *slog.Logger {
	return NewLoggerWithLevel("info")
}

func NewLoggerWithLevel(level string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: MapLogLevel(level)}))
}

func MapLogLevel(level string) slog.Level {
	level = strings.ToLower(level)

	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
