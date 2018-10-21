package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wsxiaoys/terminal/color"
)

func prepareParams() (*lingualeoArgs, error) {
	args := prepareArgs()
	err := checkConfig(&args)
	if err != nil {
		return nil, err
	}
	configArgs, err := readConfigs(args.Config)
	if err != nil {
		return nil, err
	}
	args = *mergeConfigs(&args, configArgs)
	err = checkArgs(&args)
	if err != nil {
		return nil, err
	}
	if args.Sound {
		if len(args.Player) == 0 {
			fmt.Println("Please set player parameter")
			args.Sound = false
		} else if !isCommandAvailable(args.Player) {
			fmt.Printf("Executable file %s is not availabe on your system\n", args.Player)
			args.Sound = false
		}
	}
	return &args, nil
}

func main() {
	args, err := prepareParams()
	failIfError(err)
	client, err := prepareClient()
	failIfError(err)
	auth(args, client)

	var resultsToAdd []lingualeoResult
	var urls []string
	ctx, done := context.WithCancel(context.Background())
	defer done()

	for res := range orDone(ctx, getWords(args.Words, client)) {
		res, _ := res.(result)
		if res.Error != nil {
			fmt.Println(res.Error)
			continue
		}
		if len(res.Result.Words) == 0 {
			color.Printf("@{r}There are no translations for word: @{g}['%s']\n", res.Result.Word)
			continue
		}
		printTranslate(res.Result)
		if args.Sound {
			urls = append(urls, res.Result.SoundURL)
		}
		if args.Add && (!bool(res.Result.Exists) || args.Force) {
			if len(args.Translate) > 0 {
				// Custom translation
				res.Result.Words = args.Translate
			}
			resultsToAdd = append(resultsToAdd, *res.Result)
		}
	}

	if len(urls) > 0 {
		results := make([]string, len(urls))
		for res := range orDone(ctx, downloadFiles(urls...)) {
			res, _ := res.(resultFile)
			if res.Error != nil {
				fmt.Println(res.Error)
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
				fmt.Println(err)
			}
			err = os.Remove(filename)
			if err != nil {
				fmt.Println(err)
			}
		}

	}

	if len(resultsToAdd) > 0 {
		for res := range orDone(ctx, addWords(resultsToAdd, client)) {
			res, _ := res.(result)
			if res.Error != nil {
				fmt.Println(res.Error)
				continue
			}
			printAddTranslate(res.Result)
		}
	}
}
