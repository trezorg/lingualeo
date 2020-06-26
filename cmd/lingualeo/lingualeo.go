package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/trezorg/lingualeo/pkg/api"
	"github.com/trezorg/lingualeo/pkg/files"
	"github.com/trezorg/lingualeo/pkg/translator"
	"github.com/trezorg/lingualeo/pkg/utils"
)

var version = "0.0.1"

func main() {
	args, err := translator.NewLingualeo(version)
	utils.FailIfError(err)

	ctx, done := context.WithCancel(context.Background())
	defer done()

	var wg sync.WaitGroup

	wg.Add(1)
	soundChan, addWordChan, resultsChan := args.Process(ctx, &wg)
	wg.Add(1)
	if args.DownloadSoundFile {
		go args.DownloadAndPronounce(ctx, soundChan, &wg, files.NewFileDownloader)
	} else {
		go args.Pronounce(ctx, soundChan, &wg)
	}
	wg.Add(1)
	go args.AddToDictionary(ctx, addWordChan, &wg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for sig := range stop {
			utils.PrintColorStringF("r", "Got OS signal: %s", sig)
			done()
			return
		}
	}()

	for result := range api.OrResultDone(ctx, resultsChan) {
		result.PrintTranslation()
	}

	wg.Wait()
}
