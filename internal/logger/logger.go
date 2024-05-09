package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var levelVar = new(slog.LevelVar)

func ParseLevel(level string) (slog.Level, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARNING", "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelError, fmt.Errorf("unknown log level: %s", level)
	}
}

func SetLevel(level slog.Level) {
	levelVar.Set(level)
}

func Prepare(level slog.Level) {
	SetLevel(level)
	SetHandler(DebugHandler())
}

func SetHandler(handler slog.Handler) {
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func DefaultHandler() slog.Handler {
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelVar,
	})
}

func DebugHandler() slog.Handler {
	return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
}
