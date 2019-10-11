package main

import (
	"context"
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

	results := translateWords(ctx, args, client)
	showTranslateResults(results)

	if args.Sound {
		soundUrls := getSoundUrls(results)
		if len(soundUrls) > 0 {
			playTranslateFile(ctx, args, soundUrls...)
		}
	}

	if args.Add {
		resultsToAdd := prepareResultsToAdd(results, args)
		if len(resultsToAdd) > 0 {
			addTranslationToDictionary(ctx, client, resultsToAdd)
		}
	}
}
