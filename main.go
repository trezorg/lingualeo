package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	args, err := prepareParams()
	failIfError(err)
	initLogger(args.LogLevel, args.LogPrettyPrint)
	client, err := prepareClient()
	failIfError(err)
	err = auth(args, client)
	failIfError(err)

	ctx, done := context.WithCancel(context.Background())
	defer done()

	var wg sync.WaitGroup

	wg.Add(1)
	soundChan, addWordChan, resultsChan := processTranslation(ctx, client, args, &wg)
	wg.Add(1)
	go playTranslateFiles(ctx, args, soundChan, &wg)
	wg.Add(1)
	go addTranslationToDictionary(ctx, client, addWordChan, &wg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for sig := range stop {
			printColorStringF("r", "Got OS signal: %s", sig)
			done()
			return
		}
	}()

	for result := range orDone(ctx, resultsChan) {
		printTranslate(result.(lingualeoResult))
	}

	wg.Wait()
}
