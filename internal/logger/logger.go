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

func Prepare(level slog.Level, pretty bool) {
	levelVar.Set(level)

	var handler slog.Handler
	if pretty {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: levelVar})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: levelVar})
	}

	slog.SetDefault(slog.New(handler))
}
