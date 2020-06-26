package translator

import (
	"context"
	"os"
	"sync"

	"github.com/trezorg/lingualeo/pkg/api"
	"github.com/trezorg/lingualeo/pkg/channel"
	"github.com/trezorg/lingualeo/pkg/files"
	"github.com/trezorg/lingualeo/pkg/logger"
	"github.com/trezorg/lingualeo/pkg/utils"

	"github.com/wsxiaoys/terminal/color"
)

func (args *Lingualeo) checkMediaPlayer() {
	if !args.Sound {
		return
	}
	if len(args.Player) == 0 {
		_, err := color.Printf("@{r}Please set player parameter\n", args.Player)
		if err != nil {
			logger.Log.Debug(err)
		}
		args.Sound = false
	} else if !utils.IsCommandAvailable(args.Player) {
		_, err := color.Printf("@{r}Executable file %s is not available on your system\n", args.Player)
		if err != nil {
			logger.Log.Debug(err)
		}
		args.Sound = false
	}
}

func NewLingualeo(version string) (Lingualeo, error) {
	client := prepareCliArgs(version)
	err := client.checkConfig()
	if err != nil {
		return client, err
	}
	configArgs, err := fromConfigs(&client.Config)
	if err != nil {
		return client, err
	}
	client.mergeConfigs(configArgs)
	err = client.checkArgs()
	if err != nil {
		return client, err
	}
	logger.InitLogger(client.LogLevel, client.LogPrettyPrint)
	client.checkMediaPlayer()
	a, err := api.NewAPI(client.Email, client.Password, client.Debug)
	if err != nil {
		return client, err
	}
	client.API = a
	return client, nil
}

func (args *Lingualeo) translateWords(ctx context.Context) <-chan api.OpResult {
	results := make(chan api.OpResult, len(args.Words))
	input := channel.ToStringChannel(ctx, args.Words...)
	go func() {
		defer close(results)
		for res := range api.OrOpResultDone(ctx, args.API.TranslateWords(ctx, input)) {
			if res.Error != nil {
				_, err := color.Printf("@{r}%s\n", utils.Capitalize(res.Error.Error()))
				if err != nil {
					logger.Log.Error(err)
				}
				continue
			}
			if len(res.Result.Words) == 0 {
				_, err := color.Printf("@{r}There are no translations for word: @{g}['%s']\n", res.Result.Word)
				if err != nil {
					logger.Log.Error(err)
				}
				continue
			}
			results <- res
		}
	}()
	return results
}

func (args *Lingualeo) prepareResultToAdd(result *api.Result) bool {
	if !result.Exists || args.Force {
		// Custom translation
		if len(args.Translate) > 0 {
			if args.TranslateReplace {
				result.Words = utils.Unique(args.Translate)
			} else {
				result.Words = utils.Unique(append(result.Words, args.Translate...))
			}
		}
		return true
	}
	return false
}

func (args *Lingualeo) downloadAndPronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup, downloader files.NewDownloader) {
	defer wg.Done()
	fileChannel := files.OrderedChannel(files.DownloadFiles(ctx, urls, downloader), len(urls))
	for res := range files.OrFilesDone(ctx, fileChannel) {
		if res.Error != nil {
			logger.Log.Error(res.Error)
			continue
		}
		if res.Filename == "" {
			continue
		}
		err := utils.PlaySound(args.Player, res.Filename)
		if err != nil {
			logger.Log.Error(err)
		}
		err = os.Remove(res.Filename)
		if err != nil {
			logger.Log.Error(err)
		}
	}
}

func (args *Lingualeo) pronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for res := range channel.OrStringDone(ctx, urls) {
		err := utils.PlaySound(args.Player, res)
		if err != nil {
			logger.Log.Error(err)
		}
	}
}

func (args *Lingualeo) Pronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup) {
	if args.DownloadSoundFile {
		args.downloadAndPronounce(ctx, urls, wg, files.NewFileDownloader)
	} else {
		args.pronounce(ctx, urls, wg)
	}
}

func (args *Lingualeo) AddToDictionary(ctx context.Context, resultsToAdd <-chan api.Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for res := range args.API.AddWords(ctx, resultsToAdd) {
		if res.Error != nil {
			logger.Log.Error(res.Error)
			continue
		}
		res.Result.PrintAddedTranslation()
	}
}

func (args *Lingualeo) Process(ctx context.Context, wg *sync.WaitGroup) (<-chan string, <-chan api.Result, <-chan api.Result) {

	soundChan := make(chan string, len(args.Words))
	addWordChan := make(chan api.Result, len(args.Words))
	resultsChan := make(chan api.Result, len(args.Words))

	go func() {

		defer func() {
			wg.Done()
			close(soundChan)
			close(addWordChan)
			close(resultsChan)
		}()

		for result := range args.translateWords(ctx) {
			if result.Error != nil {
				logger.Log.Error(result.Error)
				continue
			}
			if args.Sound && result.Result.SoundURL != "" {
				soundChan <- result.Result.SoundURL
			}

			if args.Add {
				if resultsToAdd := args.prepareResultToAdd(&result.Result); resultsToAdd {
					addWordChan <- result.Result
				}
			}
			resultsChan <- result.Result
		}
	}()

	return soundChan, addWordChan, resultsChan

}
