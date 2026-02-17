package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trezorg/lingualeo/internal/messages"
	"github.com/trezorg/lingualeo/internal/translator"
)

var version = "0.0.1"

func main() {
	args, err := translator.New(version)
	if err != nil {
		_ = messages.Message(messages.RED, fmt.Sprintf("Error: %v\n", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	args.TranslateWithReverseRussian(ctx)
}
