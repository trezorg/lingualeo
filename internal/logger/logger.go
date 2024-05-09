package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var levelVar = new(slog.LevelVar)

func ParseLevel(level string) (slog.Level, error) {
	switch level {
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
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelVar,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
