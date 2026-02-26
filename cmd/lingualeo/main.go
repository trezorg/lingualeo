package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trezorg/lingualeo/internal/logger"
	"github.com/trezorg/lingualeo/internal/messages"
	"github.com/trezorg/lingualeo/internal/translator"
)

var version = "0.0.1"

func main() {
	args, err := translator.New(version)
	if err != nil {
		// Handle help/version as successful exit
		if errors.Is(err, translator.ErrHelpOrVersionShown) {
			os.Exit(0)
		}
		if msgErr := messages.Message(messages.RED, "Error: %v\n", err); msgErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	level, err := logger.ParseLevel(args.LogLevel)
	if err != nil {
		if msgErr := messages.Message(messages.RED, "Error: %v\n", err); msgErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
	logger.Prepare(level, args.LogPrettyPrint)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	args.TranslateWithReverseRussian(ctx)
}
