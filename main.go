package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"

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

func fixSoundURL(rawURL string) (*string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	path := parsedURL.Path
	pathParts := strings.Split(path, "/")
	pathParts = insertIntoSlice(pathParts, 2, "3")
	parsedURL.Path = strings.Join(pathParts, "/")
	result := parsedURL.String()
	return &result, nil
}

func getSoundUrls(results *[]translateResult) []string {
	var soundUrls []string
	for _, res := range *results {
		soundUrl, err := fixSoundURL(res.Result.SoundURL)
		if err != nil {
			log.Errorf("Cannot fix sound url: %s. %#v", res.Result.SoundURL, err)
			soundUrl = &res.Result.SoundURL
		}
		soundUrls = append(soundUrls, *soundUrl)
	}
	return soundUrls
}

func getResultsToAdd(results *[]translateResult, args *lingualeoArgs) []lingualeoResult {
	var resultsToAdd []lingualeoResult
	for _, res := range *results {
		if !bool(res.Result.Exists) || args.Force {
			if len(args.Translate) > 0 {
				// Custom translation
				res.Result.Words = args.Translate
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
		resultsToAdd := getResultsToAdd(results, args)
		if len(resultsToAdd) > 0 {
			addTranslationToDictionary(ctx, client, resultsToAdd)
		}
	}
}
