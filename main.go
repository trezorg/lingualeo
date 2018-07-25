package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type lingualeoWordResult struct {
	Votes int    `json:"votes"`
	Value string `json:"value"`
}

type result struct {
	Error  error
	Result *lingualeoResult
}

type resultFile struct {
	Error    error
	Filename string
	Index    int
}

type lingualeoResult struct {
	Word          string                `json:"-"`
	Words         []string              `json:"-"`
	Exists        convertibleBoolean    `json:"is_user"`
	SoundURL      string                `json:"sound_url"`
	Transcription string                `json:"transcription"`
	Translate     []lingualeoWordResult `json:"translate"`
	ErrorMsg      string                `json:"error_msg"`
}

type lingualeoArgs struct {
	Email    string
	Password string
	Config   string
	Player   string
	Words    []string
	Force    bool
	Add      bool
	Sound    bool
	Debug    bool
}

func failIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func playSound(player string, url string) error {
	cmd := exec.Command(player, url)
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

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
		parseAndSortTranslate(res.Result)
		if len(res.Result.Words) == 0 {
			continue
		}
		printTranslate(res.Result)
		if args.Sound {
			urls = append(urls, res.Result.SoundURL)
		}
		if args.Add && (!bool(res.Result.Exists) || args.Force) {
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

			var subTitle string
			if res.Result.Exists {
				subTitle = "Updated existing"
			} else {
				subTitle = "Added new"
			}

			printColorString("r", fmt.Sprintf("%s word %s", subTitle, res.Result.Word))
		}
	}
}
