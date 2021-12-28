package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/trezorg/lingualeo/pkg/messages"
	"github.com/trezorg/lingualeo/pkg/translator"
	"github.com/trezorg/lingualeo/pkg/utils"
)

var version = "0.0.1"

func main() {
	args, err := translator.NewLingualeo(version)
	utils.FailIfError(err)

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
