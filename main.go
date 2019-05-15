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
	args := prepareArgs()
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

func translate(ctx context.Context, args *lingualeoArgs, client *http.Client) ([]lingualeoResult, []string) {
	var resultsToAdd []lingualeoResult
	var soundUrls []string
	for res := range orDone(ctx, getWords(args.Words, client)) {
		res, _ := res.(result)
		if res.Error != nil {
			log.Error(res.Error)
			continue
		}
		if len(res.Result.Words) == 0 {
			_, err := color.Printf("@{r}There are no translations for word: @{g}['%s']\n", res.Result.Word)
			if err != nil {
				log.Debug(err)
			}
			continue
		}
		printTranslate(res.Result)
		if args.Sound {
			soundUrls = append(soundUrls, res.Result.SoundURL)
		}
		if args.Add && (!bool(res.Result.Exists) || args.Force) {
			if len(args.Translate) > 0 {
				// Custom translation
				res.Result.Words = args.Translate
			}
			resultsToAdd = append(resultsToAdd, *res.Result)
		}
	}
	return resultsToAdd, soundUrls
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
		res, _ := res.(result)
		if res.Error != nil {
			log.Error(res.Error)
			continue
		}
		printAddedTranslation(res.Result)
	}
}

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

	resultsToAdd, soundUrls := translate(ctx, args, client)

	if len(soundUrls) > 0 {
		playTranslateFile(ctx, args, soundUrls...)
	}

	if len(resultsToAdd) > 0 {
		addTranslationToDictionary(ctx, client, resultsToAdd)
	}
}
