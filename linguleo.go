package main

import (
	"context"
	"net/http"
	"os"
	"sync"

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

func translateWords(ctx context.Context, args *lingualeoArgs, client *http.Client) <-chan interface{} {
	results := make(chan interface{}, len(args.Words))
	go func() {
		defer close(results)
		for res := range orDone(ctx, getWords(args.Words, client)) {
			res, _ := res.(translateResult)
			if res.Error != nil {
				_, err := color.Printf("@{r}%s\n", capitalize(res.Error.Error()))
				if err != nil {
					log.Error(err)
				}
				continue
			}
			if len(res.Result.Words) == 0 {
				_, err := color.Printf("@{r}There are no translations for word: @{g}['%s']\n", res.Result.Word)
				if err != nil {
					log.Error(err)
				}
				continue
			}
			results <- res
		}
	}()
	return results
}

func prepareResultToAdd(result lingualeoResult, args *lingualeoArgs) *lingualeoResult {
	if !result.Exists || args.Force {
		// Custom translation
		if len(args.Translate) > 0 {
			if args.TranslateReplace {
				result.Words = unique(args.Translate)
			} else {
				result.Words = unique(append(result.Words, args.Translate...))
			}
		}
		return &result
	}
	return nil
}

func playTranslateDownloadFiles(ctx context.Context, args *lingualeoArgs, urls <-chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for res := range orDone(ctx, orderedChannel(downloadFiles(ctx, urls), len(urls))) {
		res, _ := res.(resultFile)
		if res.Error != nil {
			log.Error(res.Error)
			continue
		}
		if res.Filename == "" {
			continue
		}
		err := playSound(args.Player, res.Filename)
		if err != nil {
			log.Error(err)
		}
		err = os.Remove(res.Filename)
		if err != nil {
			log.Error(err)
		}
	}
}

func playTranslateFiles(ctx context.Context, args *lingualeoArgs, urls <-chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for res := range orDone(ctx, urls) {
		err := playSound(args.Player, res.(string))
		if err != nil {
			log.Error(err)
		}
	}
}

func addTranslationToDictionary(ctx context.Context, client *http.Client, resultsToAdd <-chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for res := range addWords(ctx, resultsToAdd, client) {
		res, _ := res.(translateResult)
		if res.Error != nil {
			log.Error(res.Error)
			continue
		}
		printAddedTranslation(res.Result)
	}
}

func processTranslation(ctx context.Context, client *http.Client, args *lingualeoArgs, wg *sync.WaitGroup) (_, _, _ <-chan interface{}) {

	soundChan := make(chan interface{}, len(args.Words))
	addWordChan := make(chan interface{}, len(args.Words))
	resultsChan := make(chan interface{}, len(args.Words))

	go func() {

		defer func() {
			wg.Done()
			close(soundChan)
			close(addWordChan)
			close(resultsChan)
		}()

		for value := range orDone(ctx, translateWords(ctx, args, client)) {
			result, _ := value.(translateResult)
			if args.Sound && result.Result.SoundURL != "" {
				soundChan <- result.Result.SoundURL
			}

			if args.Add {
				if resultsToAdd := prepareResultToAdd(result.Result, args); resultsToAdd != nil {
					addWordChan <- *resultsToAdd
				}
			}
			resultsChan <- result.Result
		}
	}()

	return soundChan, addWordChan, resultsChan

}
