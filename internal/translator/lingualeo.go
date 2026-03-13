package translator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/trezorg/lingualeo/internal/api"
	"github.com/trezorg/lingualeo/internal/channel"
	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/internal/messages"
	"github.com/trezorg/lingualeo/internal/slice"
	"github.com/trezorg/lingualeo/internal/visualizer/browser"
	"github.com/trezorg/lingualeo/internal/visualizer/term"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var errUnknownVisualiseType = errors.New("unknown visualize type")

type Lingualeo struct {
	// Embedded interfaces (dependencies)
	api.Client `json:"-" yaml:"-" toml:"-"`
	Downloader `json:"-" yaml:"-" toml:"-"`
	Pronouncer `json:"-" yaml:"-" toml:"-"`
	Outputer   `json:"-" yaml:"-" toml:"-"`

	// Embedded config - inline tags preserve flat access for config file parsing
	//nolint:revive // inline tags required for yaml/toml/json v2 embedding
	Config `yaml:",inline" json:",inline" toml:",inline"`

	// Runtime inputs (not serialized)
	ConfigPath  string   // Path to config file (renamed from Config to avoid collision)
	Words       []string // Words to translate
	Translation []string // Custom translation override
}

func visualizer(vt VisualiseType) (Visualizer, error) {
	switch vt {
	case Default:
		return browser.New(), nil
	case Term:
		if term.Mode() == term.Unknown {
			return browser.New(), nil
		}
		return term.New(), nil
	default:
		return nil, fmt.Errorf("%w: %s", errUnknownVisualiseType, vt)
	}
}

func outputer(visualize bool, vt VisualiseType) (Outputer, error) {
	if !visualize {
		return Output{}, nil
	}
	viz, err := visualizer(vt)
	if err != nil {
		return nil, err
	}
	return OutputVisualizer{Visualizer: viz}, nil
}

// NewOutputer creates an outputer based on visualize setting.
// Exported for use in main.go for explicit dependency injection.
func NewOutputer(visualize bool, vt VisualiseType) (Outputer, error) {
	return outputer(visualize, vt)
}

// sendOperationResult sends a result to the output channel.
// The return value is intentionally ignored as this is a fire-and-forget
// operation - if the context is cancelled, the result is simply dropped.
// This is by design: when the pipeline is shutting down, we don't want to
// block on sending results that won't be processed anyway.
func sendOperationResult(ctx context.Context, out chan<- api.OperationResult, res api.OperationResult) {
	_ = sendToChanWithContext(ctx, out, res)
}

func sendToChanWithContext[T any](ctx context.Context, out chan<- T, value T) bool {
	if ctx.Err() != nil {
		return false
	}
	select {
	case <-ctx.Done():
		return false
	case out <- value:
		return true
	}
}

func workerCount(workers int) int {
	if workers <= 0 {
		return defaultWorkers
	}

	return workers
}

func workerCountForItems(workers int, items int) int {
	count := workerCount(workers)
	if items > 0 && count > items {
		return items
	}

	return count
}

// Parse parses CLI args and config files, returning config without initializing dependencies.
// Use functional options to inject dependencies after calling Parse.
func Parse(version string, options ...Option) (Lingualeo, error) {
	client, err := prepareTranslator(version)
	if err != nil {
		return client, err
	}
	if err = client.promptPasswordIfNeeded(); err != nil {
		return client, err
	}
	if err = client.checkArgs(); err != nil {
		return client, err
	}
	if client.Debug {
		client.LogLevel = "DEBUG"
	}
	if err = client.checkMediaPlayer(); err != nil {
		slog.Warn("media player check failed", "error", err)
	}

	for _, option := range options {
		if err = option(&client); err != nil {
			return client, err
		}
	}
	return client, nil
}

// Validate ensures required dependencies are set.
// Call this after applying all options with dependency injection.
func (l *Lingualeo) Validate() error {
	var err error
	if l.Client == nil {
		err = errors.Join(err, errMissingClient)
	}
	if l.Outputer == nil {
		err = errors.Join(err, errMissingOutputer)
	}
	if l.Sound && l.Pronouncer == nil {
		err = errors.Join(err, errMissingPronouncer)
	}
	if l.Sound && l.DownloadSoundFile && l.Downloader == nil {
		err = errors.Join(err, errMissingDownloader)
	}

	return err
}

// translateWords translate words from string channel
func translateWords(ctx context.Context, translator api.Client, results <-chan string, workers int) <-chan api.OperationResult {
	out := make(chan api.OperationResult)
	var wg sync.WaitGroup
	for range workerCount(workers) {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case word, ok := <-results:
					if !ok {
						return
					}
					sendOperationResult(ctx, out, translator.TranslateWord(ctx, word))
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// addWords add words
func addWords(ctx context.Context, translator api.Client, results <-chan api.Result, workers int) <-chan api.OperationResult {
	out := make(chan api.OperationResult)
	var wg sync.WaitGroup
	for range workerCount(workers) {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case res, ok := <-results:
					if !ok {
						return
					}
					for _, translate := range res.AddWords {
						added := translator.AddWord(ctx, res.Word, translate)
						added.Result.AddWords = []string{translate}
						sendOperationResult(ctx, out, added)
					}
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (l *Lingualeo) checkMediaPlayer() error {
	if !l.Sound {
		return nil
	}
	if len(l.Player) == 0 {
		l.Sound = false
		return errors.New("player parameter not set, sound disabled")
	}
	if !isCommandAvailable(l.Player) {
		l.Sound = false
		return fmt.Errorf("player executable %s not available, sound disabled", l.Player)
	}
	return nil
}

func (l *Lingualeo) translateWords(ctx context.Context, words []string) <-chan api.OperationResult {
	results := make(chan api.OperationResult, len(words))
	input := channel.ToChannel(ctx, words...)
	workers := workerCountForItems(l.Workers, len(words))
	go func() {
		defer close(results)
		ch := translateWords(ctx, l.Client, input, workers)
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
			if len(res.Result.Translate) == 0 {
				_ = messages.Message(messages.RED, "There are no translations for word: ")
				err := messages.Message(messages.GREEN, "['%s']\n", res.Result.Word)
				if err != nil {
					slog.Error("cannot show message", "error", err)
				}
				continue
			}
			if !sendToChanWithContext(ctx, results, res) {
				return
			}
		}
	}()
	return results
}

func (l *Lingualeo) prepareResultToAdd(result *api.Result) bool {
	// Custom translation
	if len(l.Translation) > 0 {
		result.SetTranslation(l.Translation)
		return true
	}
	// Use top translation from API if available
	if len(result.Translate) > 0 {
		translations := make([]string, 0, len(result.Translate))
		for _, item := range result.Translate {
			translations = append(translations, item.Value)
		}
		result.SetTranslation(translations)
		return true
	}
	return false
}

func (l *Lingualeo) downloadAndPronounce(ctx context.Context, urls <-chan string, wordCount int) {
	workers := workerCountForItems(l.Workers, wordCount)
	fileChannel := files.OrderedChannel(downloadFiles(ctx, urls, l.Downloader, workers), wordCount)
	for res := range channel.OrDone(ctx, fileChannel) {
		if res.Error != nil {
			slog.Error("cannot download", "error", res.Error)
			continue
		}
		if res.Filename == "" {
			continue
		}
		if err := l.Play(ctx, res.Filename); err != nil {
			slog.Error("cannot play filename", "filename", res.Filename, "error", err)
		}
		if err := l.Remove(res.Filename); err != nil {
			slog.Error("cannot remove filename", "filename", res.Filename, "error", err)
		}
	}
}

func (l *Lingualeo) playURLs(ctx context.Context, urls <-chan string) {
	for url := range channel.OrDone(ctx, urls) {
		if err := l.Play(ctx, url); err != nil {
			slog.Error("cannot play url", "url", url, "error", err)
		}
	}
}

// Pronounce downloads and pronounce words
func (l *Lingualeo) Pronounce(ctx context.Context, urls <-chan string, wordCount int) {
	if l.DownloadSoundFile {
		l.downloadAndPronounce(ctx, urls, wordCount)
		return
	}
	l.playURLs(ctx, urls)
}

func (l *Lingualeo) AddToDictionary(ctx context.Context, resultsToAdd <-chan api.Result, wordCount int) {
	workers := workerCountForItems(l.Workers, wordCount)
	ch := addWords(ctx, l.Client, resultsToAdd, workers)
	for res := range ch {
		if res.Error != nil {
			slog.Error("cannot add word to dictionary", "word", res.Result.Word, "error", res.Error)
			continue
		}
		if err := PrintAddedTranslation(res.Result); err != nil {
			slog.Error("cannot print added translation", "word", res.Result.Word, "error", err)
		}
	}
}

type Channels struct {
	sound   <-chan string
	add     <-chan api.Result
	results <-chan api.Result
}

func (l *Lingualeo) Process(ctx context.Context, words []string, wg *sync.WaitGroup) Channels {
	soundChan := make(chan string, len(words))
	addWordChan := make(chan api.Result, len(words))
	resultsChan := make(chan api.Result, len(words))

	go func() {
		defer func() {
			wg.Done()
			close(soundChan)
			close(addWordChan)
			close(resultsChan)
		}()

		for result := range l.translateWords(ctx, words) {
			if result.Error != nil {
				slog.Error("cannot translate word", "word", result.Result.Word, "error", result.Error)
				continue
			}
			if l.Sound {
				if !sendToChanWithContext(ctx, soundChan, result.Result.SoundURL) {
					return
				}
			}

			if l.Add {
				if resultsToAdd := l.prepareResultToAdd(&result.Result); resultsToAdd {
					if !sendToChanWithContext(ctx, addWordChan, result.Result) {
						return
					}
				}
			}
			if !sendToChanWithContext(ctx, resultsChan, result.Result) {
				return
			}
		}
	}()

	return Channels{
		sound:   soundChan,
		add:     addWordChan,
		results: resultsChan,
	}
}

func (l *Lingualeo) translateToChan(ctx context.Context, words []string) <-chan api.Result {
	var wg sync.WaitGroup
	wg.Add(1)
	channels := l.Process(ctx, words, &wg)
	if l.Sound {
		wg.Go(func() {
			l.Pronounce(ctx, channels.sound, len(words))
		})
	}
	if l.Add {
		wg.Go(func() {
			l.AddToDictionary(ctx, channels.add, len(words))
		})
	}

	ch := make(chan api.Result, len(words))

	go func() {
		defer close(ch)
		for result := range channel.OrDone(ctx, channels.results) {
			if !sendToChanWithContext(ctx, ch, result) {
				return
			}
		}
		wg.Wait()
	}()

	return ch
}

func (l *Lingualeo) translateAndOutput(ctx context.Context, words []string, collectReverse bool) ([]string, error) {
	englishWords := make([]string, 0, len(words))
	for result := range channel.OrDone(ctx, l.translateToChan(ctx, words)) {
		if err := l.Output(ctx, result); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, err
			}
			slog.Error("cannot translate word", "word", result.Word, "error", err)
		}
		for _, word := range result.Translate {
			if collectReverse && isEnglishWord(word.Value) {
				englishWords = append(englishWords, word.Value)
			}
		}
	}

	return slice.Unique(englishWords), nil
}

func (l *Lingualeo) TranslateWithReverseRussian(ctx context.Context) {
	englishWords, err := l.translateAndOutput(ctx, l.Words, l.ReverseTranslate)
	if err != nil || len(englishWords) == 0 {
		return
	}
	_, err = l.translateAndOutput(ctx, englishWords, false)
	if err != nil {
		return
	}
}
