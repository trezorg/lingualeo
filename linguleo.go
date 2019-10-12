package main

import (
	"context"
	"net/http"
	"os"

	"github.com/wsxiaoys/terminal/color"
)

func checkMediaPlayer(args *lingualeoArgs) {
	if len(args.Player) == 0 {
		_, err := color.Printf("@{r}Please set player parameter\n", args.Player)
		if err != nil {
			log.Debug(err)
		}
		args.Sound = false
	} else if !isCommandAvailable(args.Player) {
		_, err := color.Printf("@{r}Executable file %s is not available on your system\n", args.Player)
		if err != nil {
			log.Debug(err)
		}
		args.Sound = false
	}
}

func prepareParams() (*lingualeoArgs, error) {
	args := prepareCliArgs()
	err := checkConfig(&args)
	if err != nil {
		return nil, err
	}
	configArgs, err := readConfigs(&args.Config)
	if err != nil {
		return nil, err
	}
	args = *mergeConfigs(&args, configArgs)
	err = checkArgs(&args)
	if err != nil {
		return nil, err
	}
	if args.Sound {
		checkMediaPlayer(&args)
	}
	return &args, nil
}

func translateWords(ctx context.Context, args *lingualeoArgs, client *http.Client) *[]translateResult {
	var results []translateResult
	for res := range orDone(ctx, getWords(args.Words, client)) {
		res, _ := res.(translateResult)
		if res.Error != nil {
			_, err := color.Printf("@{r}%s\n", capitalize(res.Error.Error()))
			if err != nil {
				log.Debug(err)
			}
			continue
		}
		if len(res.Result.Words) == 0 {
			_, err := color.Printf("@{r}There are no translations for word: @{g}['%s']\n", res.Result.Word)
			if err != nil {
				log.Debug(err)
			}
			continue
		}
		results = append(results, res)
	}
	return &results
}

func showTranslateResults(results *[]translateResult) {
	for _, res := range *results {
		printTranslate(res.Result)
	}
}

func getSoundUrls(results *[]translateResult) []string {
	var soundUrls []string
	for _, res := range *results {
		soundUrls = append(soundUrls, res.Result.SoundURL)
	}
	return soundUrls
}

func prepareResultsToAdd(results *[]translateResult, args *lingualeoArgs) []lingualeoResult {
	var resultsToAdd []lingualeoResult
	for _, res := range *results {
		if !res.Result.Exists || args.Force {
			if len(args.Translate) > 0 {
				// Custom translation
				if args.TranslateReplace {
					res.Result.Words = unique(args.Translate)
				} else {
					res.Result.Words = unique(append(res.Result.Words, args.Translate...))
				}
			}
			resultsToAdd = append(resultsToAdd, *res.Result)
		}
	}
	return resultsToAdd
}

func playTranslateFile(ctx context.Context, args *lingualeoArgs, urls ...string) {
	results := make([]string, len(urls))
	for res := range orDone(ctx, downloadFiles(urls...)) {
		res, _ := res.(resultFile)
		if res.Error != nil {
			log.Error(res.Error)
			continue
		}
		results[res.Index] = res.Filename
	}
	for _, filename := range results {
		if filename == "" {
			continue
		}
		err := playSound(args.Player, filename)
		if err != nil {
			log.Error(err)
		}
		err = os.Remove(filename)
		if err != nil {
			log.Error(err)
		}
	}
}

func addTranslationToDictionary(ctx context.Context, client *http.Client, resultsToAdd []lingualeoResult) {
	for res := range orDone(ctx, addWords(resultsToAdd, client)) {
		res, _ := res.(translateResult)
		if res.Error != nil {
			log.Error(res.Error)
			continue
		}
		printAddedTranslation(res.Result)
	}
}
