package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trezorg/lingualeo/pkg/messages"
	"github.com/trezorg/lingualeo/pkg/translator"
)

var version = "0.0.1"

func main() {
	args, err := translator.New(version)
	if err != nil {
		_ = messages.Message(messages.RED, fmt.Sprintf("Error: %v\n", err))
		os.Exit(1)
	}

	ctx, done := context.WithCancel(context.Background())
	defer done()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for sig := range stop {
			_ = messages.Message(messages.RED, "Got OS signal: %s\n", sig)
			done()
			return
		}
	}()

	args.TranslateWithReverseRussian(ctx, translator.ProcessResultImpl)

}
