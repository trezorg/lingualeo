package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/trezorg/lingualeo/internal/logger"
	"github.com/trezorg/lingualeo/internal/messages"
	"github.com/trezorg/lingualeo/internal/translator"
)

var version = "0.0.1"

func main() {
	os.Exit(run())
}

func run() int {
	args, err := translator.New(version)
	if err != nil {
		// Handle help/version as successful exit
		if errors.Is(err, translator.ErrHelpOrVersionShown) {
			return 0
		}
		if msgErr := messages.Message(messages.RED, "Error: %v\n", err); msgErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return 1
	}

	level, err := logger.ParseLevel(args.LogLevel)
	if err != nil {
		if msgErr := messages.Message(messages.RED, "Error: %v\n", err); msgErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return 1
	}
	logger.Prepare(level, args.LogPrettyPrint)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err = args.Auth(ctx); err != nil {
		slog.ErrorContext(ctx, "auth error", "error", err)
		return 1
	}

	args.TranslateWithReverseRussian(ctx)
	return 0
}
