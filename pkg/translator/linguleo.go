package translator

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/pkg/api"
	"github.com/trezorg/lingualeo/pkg/channel"

	"github.com/trezorg/lingualeo/internal/logger"
	"github.com/trezorg/lingualeo/pkg/messages"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (args *Lingualeo) checkMediaPlayer() {
	if !args.Sound {
		return
	}
	if len(args.Player) == 0 {
		err := messages.Message(messages.RED, "Please set player parameter\n")
		if err != nil {
			slog.Error("cannot show message", "error", err)
		}
		args.Sound = false
	} else if !isCommandAvailable(args.Player) {
		err := messages.Message(messages.RED, "Executable file %s is not available on your system\n", args.Player)
		if err != nil {
			slog.Error("cannot show message", "error", err)
		}
		args.Sound = false
	}
}

// Translator interface
//
//go:generate mockery
type Translator interface {
	TranslateWord(word string) api.OperationResult
	AddWord(word string, translate string) api.OperationResult
}

type Lingualeo struct {
	Translator        Translator
	LogLevel          string `yaml:"log_level" json:"log_level" toml:"log_level"`
	Password          string `yaml:"password" json:"password" toml:"password"`
	Config            string
	Player            string `yaml:"player" json:"player" toml:"player"`
	Email             string `yaml:"email" json:"email" toml:"email"`
	Words             []string
	Translation       []string
	Add               bool `yaml:"add" json:"add" toml:"add"`
	Sound             bool `yaml:"sound" json:"sound" toml:"sound"`
	Debug             bool `yaml:"debug" json:"debug" toml:"debug"`
	DownloadSoundFile bool `yaml:"download" json:"download" toml:"download"`
	LogPrettyPrint    bool `yaml:"log_pretty_print" json:"log_pretty_print" toml:"log_pretty_print"`
	ReverseTranslate  bool `yaml:"reverse_translate" json:"reverse_translate" toml:"reverse_translate"`
}

// New initialize lingualeo client
func New(version string) (Lingualeo, error) {
	client, err := prepareArgs(version)
	if err != nil {
		return client, err
	}
	err = client.checkConfig()
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
	if client.Debug {
		client.LogLevel = "DEBUG"
		client.LogPrettyPrint = true
	}
	level, err := logger.ParseLevel(client.LogLevel)
	if err != nil {
		return client, err
	}
	logger.Prepare(level)
	client.checkMediaPlayer()
	client.Translator, err = api.New(client.Email, client.Password, client.Debug)
	if err != nil {
		return client, err
	}
	return client, nil
}

// translateWords translate words from string channel
func translateWords(ctx context.Context, translator Translator, results <-chan string) <-chan api.OperationResult {
	out := make(chan api.OperationResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for word := range channel.OrDone(ctx, results) {
			wg.Add(1)
			go func(word string) {
				defer wg.Done()
				out <- translator.TranslateWord(word)
			}(word)
		}
	}()
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

// addWords add words
func addWords(ctx context.Context, translator Translator, results <-chan api.Result) <-chan api.OperationResult {
	out := make(chan api.OperationResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for res := range channel.OrDone(ctx, results) {
			for _, translate := range res.AddWords {
				wg.Add(1)
				result := res
				go func(word, transate string) {
					defer wg.Done()
					added := translator.AddWord(word, transate)
					added.Result.AddWords = []string{transate}
					out <- added
				}(result.Word, translate)
			}
		}
	}()
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func (args *Lingualeo) translateWords(ctx context.Context) <-chan api.OperationResult {
	results := make(chan api.OperationResult, len(args.Words))
	input := channel.ToChannel(ctx, args.Words...)
	go func() {
		defer close(results)
		ch := translateWords(ctx, args.Translator, input)
		for res := range channel.OrDone(ctx, ch) {
			if res.Error != nil {
				err := messages.Message(
					messages.RED,
					"%s\n",
					cases.Title(language.Make(strings.ToLower(res.Error.Error()))),
				)
				if err != nil {
					slog.Error("cannot show message", "error", err)
				}
				continue
			}
			if len(res.Result.Words) == 0 {
				_ = messages.Message(messages.RED, "There are no translations for word: ")
				err := messages.Message(messages.GREEN, "['%s']\n", res.Result.Word)
				if err != nil {
					slog.Error("cannot show message", "error", err)
				}
				continue
			}
			results <- res
		}
	}()
	return results
}

func (args *Lingualeo) prepareResultToAdd(result *api.Result) bool {
	// Custom translation
	if len(args.Translation) > 0 {
		result.SetTranslation(args.Translation)
		return true
	}
	return false
}

func (args *Lingualeo) downloadAndPronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup, downloader Downloader) {
	defer wg.Done()
	fileChannel := files.OrderedChannel(DownloadFiles(ctx, urls, downloader), len(urls))
	for res := range channel.OrDone(ctx, fileChannel) {
		if res.Error != nil {
			slog.Error("cannot download", "error", res.Error)
			continue
		}
		if res.Filename == "" {
			continue
		}
		err := PlaySound(args.Player, res.Filename)
		if err != nil {
			slog.Error("cannot play filename", "filename", res.Filename, "error", err)
		}
		err = os.Remove(res.Filename)
		if err != nil {
			slog.Error("cannot remove filename", "filename", res.Filename, "error", err)
		}
	}
}

func (args *Lingualeo) pronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range channel.OrDone(ctx, urls) {
		err := PlaySound(args.Player, url)
		if err != nil {
			slog.Error("cannot play url", "url", url, "error", err)
		}
	}
}

// Pronounce downloads and pronounce words
func (args *Lingualeo) Pronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup) {
	if args.DownloadSoundFile {
		args.downloadAndPronounce(ctx, urls, wg, &files.FileDownloader{})
	} else {
		args.pronounce(ctx, urls, wg)
	}
}

// AddToDictionary adds words to dictionary
func (args *Lingualeo) AddToDictionary(ctx context.Context, resultsToAdd <-chan api.Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := addWords(ctx, args.Translator, resultsToAdd)
	for res := range ch {
		if res.Error != nil {
			slog.Error("cannot add word to dictionary", "word", res.Result.Word, "error", res.Error)
			continue
		}
		res.Result.PrintAddedTranslation()
	}
}

// Process starts translation process
func (args *Lingualeo) Process(ctx context.Context, wg *sync.WaitGroup) (<-chan string, <-chan api.Result, <-chan api.Result) {
	soundChan := make(chan string, len(args.Words))
	addWordChan := make(chan api.Result, len(args.Translation))
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
				slog.Error("cannot translate word", "word", result.Result.Word, "error", result.Error)
				continue
			}
			if args.Sound {
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

type processResult func(api.Result) error

func ProcessResultImpl(result api.Result) error {
	result.PrintTranslation()
	return nil
}

func (args *Lingualeo) translateToChan(ctx context.Context) <-chan api.Result {
	var wg sync.WaitGroup
	wg.Add(1)
	soundChan, addWordChan, resultsChan := args.Process(ctx, &wg)
	if args.Sound {
		wg.Add(1)
		go args.Pronounce(ctx, soundChan, &wg)
	}
	if args.Add {
		wg.Add(1)
		go args.AddToDictionary(ctx, addWordChan, &wg)
	}

	ch := make(chan api.Result, len(args.Words))

	go func() {
		defer close(ch)
		for result := range channel.OrDone(ctx, resultsChan) {
			ch <- result
		}
		wg.Wait()
	}()

	return ch
}

func (args *Lingualeo) TranslateWithReverseRussian(ctx context.Context, resultFunc processResult) {
	// TranslateWithReverseRussian translates russian words,
	// gets english translations and translates them once more
	var englishWords []string
	for result := range channel.OrDone(ctx, args.translateToChan(ctx)) {
		if err := resultFunc(result); err != nil {
			slog.Error("cannot translate word", "word", result.Word, "error", err)
		}
		for _, word := range result.Words {
			if args.ReverseTranslate && isEnglishWord(word) {
				englishWords = append(englishWords, word)
			}
		}
	}
	if len(englishWords) > 0 {
		args.Words = englishWords
		for result := range channel.OrDone(ctx, args.translateToChan(ctx)) {
			if err := resultFunc(result); err != nil {
				slog.Error("cannot translate word", "word", result.Word, "error", err)
			}
		}
	}
}
