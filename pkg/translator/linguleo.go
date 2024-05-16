package translator

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"

	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/internal/visualizer/browser"
	"github.com/trezorg/lingualeo/internal/visualizer/term"
	"github.com/trezorg/lingualeo/pkg/api"
	"github.com/trezorg/lingualeo/pkg/channel"

	"github.com/trezorg/lingualeo/internal/logger"
	"github.com/trezorg/lingualeo/internal/player"
	"github.com/trezorg/lingualeo/pkg/messages"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Translator interface
//
//go:generate mockery
type Translator interface {
	TranslateWord(word string) api.OperationResult
	AddWord(word string, translate string) api.OperationResult
}

type VisualiseType string

const (
	Default VisualiseType = "default"
	Term    VisualiseType = "term"
)

var (
	VisualiseTypes       = []VisualiseType{Default, Term}
	VisualiseTypeDefault = Default
)

func (v *VisualiseType) Set(value string) error {
	vt := VisualiseType(value)
	switch vt {
	case Default, Term:
		*v = vt
		return nil
	default:
		return fmt.Errorf("allowed: %s", VisualiseTypes)
	}
}

func (v *VisualiseType) String() string {
	return string(*v)
}

type Lingualeo struct {
	Translator
	Downloader
	Pronouncer
	Visualizer
	Email             string        `yaml:"email" json:"email" toml:"email"`
	VisualiseType     VisualiseType `yaml:"visualize_type" json:"visualize_type" toml:"visualize_type"`
	Config            string
	Player            string `yaml:"player" json:"player" toml:"player"`
	LogLevel          string `yaml:"log_level" json:"log_level" toml:"log_level"`
	Password          string `yaml:"password" json:"password" toml:"password"`
	Translation       []string
	Words             []string
	Add               bool `yaml:"add" json:"add" toml:"add"`
	Sound             bool `yaml:"sound" json:"sound" toml:"sound"`
	Visualise         bool `yaml:"visualize" json:"visualize" toml:"visualize"`
	Debug             bool `yaml:"debug" json:"debug" toml:"debug"`
	DownloadSoundFile bool `yaml:"download" json:"download" toml:"download"`
	LogPrettyPrint    bool `yaml:"log_pretty_print" json:"log_pretty_print" toml:"log_pretty_print"`
	ReverseTranslate  bool `yaml:"reverse_translate" json:"reverse_translate" toml:"reverse_translate"`
}

// New initialize lingualeo client
func New(version string, options ...Option) (Lingualeo, error) {
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
	if err = client.checkArgs(); err != nil {
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

	for _, option := range options {
		if err = option(&client); err != nil {
			return client, err
		}
	}
	if client.Translator == nil {
		if client.Translator, err = api.New(client.Email, client.Password, client.Debug); err != nil {
			return client, err
		}
	}
	if client.Pronouncer == nil {
		client.Pronouncer = player.New(client.Player)
	}
	if client.Downloader == nil {
		client.Downloader = files.New()
	}
	if client.Visualizer == nil {
		switch client.VisualiseType {
		case Default:
			client.Visualizer = browser.New()
		case Term:
			client.Visualizer = term.New()
		default:
			err = fmt.Errorf("unknown visualize type: %s", client.VisualiseType)
		}
	}
	return client, err
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

// visualizeWords visualize words from URL channel
func visualizeWords(ctx context.Context, visualizer Visualizer, results <-chan *url.URL) <-chan error {
	out := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for u := range channel.OrDone(ctx, results) {
			wg.Add(1)
			go func(u *url.URL) {
				defer wg.Done()
				out <- visualizer.Show(u)
			}(u)
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

func (l *Lingualeo) checkMediaPlayer() {
	if !l.Sound {
		return
	}
	if len(l.Player) == 0 {
		err := messages.Message(messages.RED, "Please set player parameter\n")
		if err != nil {
			slog.Error("cannot show message", "error", err)
		}
		l.Sound = false
	} else if !isCommandAvailable(l.Player) {
		err := messages.Message(messages.RED, "Executable file %s is not available on your system\n", l.Player)
		if err != nil {
			slog.Error("cannot show message", "error", err)
		}
		l.Sound = false
	}
}

func (l *Lingualeo) translateWords(ctx context.Context) <-chan api.OperationResult {
	results := make(chan api.OperationResult, len(l.Words))
	input := channel.ToChannel(ctx, l.Words...)
	go func() {
		defer close(results)
		ch := translateWords(ctx, l.Translator, input)
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
			results <- res
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
	return false
}

func (l *Lingualeo) downloadAndPronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	fileChannel := files.OrderedChannel(downloadFiles(ctx, urls, l.Downloader), len(urls))
	for res := range channel.OrDone(ctx, fileChannel) {
		if res.Error != nil {
			slog.Error("cannot download", "error", res.Error)
			continue
		}
		if res.Filename == "" {
			continue
		}
		if err := l.Play(res.Filename); err != nil {
			slog.Error("cannot play filename", "filename", res.Filename, "error", err)
		}
		if err := l.Remove(res.Filename); err != nil {
			slog.Error("cannot remove filename", "filename", res.Filename, "error", err)
		}
	}
}

func (l *Lingualeo) pronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range channel.OrDone(ctx, urls) {
		err := l.Play(url)
		if err != nil {
			slog.Error("cannot play url", "url", url, "error", err)
		}
	}
}

// Pronounce downloads and pronounce words
func (l *Lingualeo) Pronounce(ctx context.Context, urls <-chan string, wg *sync.WaitGroup) {
	if l.DownloadSoundFile {
		l.downloadAndPronounce(ctx, urls, wg)
	} else {
		l.pronounce(ctx, urls, wg)
	}
}

// AddToDictionary adds words to dictionary
func (l *Lingualeo) AddToDictionary(ctx context.Context, resultsToAdd <-chan api.Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := addWords(ctx, l.Translator, resultsToAdd)
	for res := range ch {
		if res.Error != nil {
			slog.Error("cannot add word to dictionary", "word", res.Result.Word, "error", res.Error)
			continue
		}
		res.Result.PrintAddedTranslation()
	}
}

// Visualize show words pictures
func (l *Lingualeo) Visualize(ctx context.Context, urls <-chan *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := visualizeWords(ctx, l.Visualizer, urls)
	for err := range ch {
		if err != nil {
			slog.Error("cannot show word picture", "error", err)
		}
	}
}

type Channels struct {
	sound     <-chan string
	visualize <-chan *url.URL
	add       <-chan api.Result
	results   <-chan api.Result
}

// Process starts translation process
func (l *Lingualeo) Process(ctx context.Context, wg *sync.WaitGroup) Channels {
	soundChan := make(chan string, len(l.Words))
	visualizeChan := make(chan *url.URL, len(l.Words))
	addWordChan := make(chan api.Result, len(l.Translation))
	resultsChan := make(chan api.Result, len(l.Words))

	go func() {
		defer func() {
			wg.Done()
			close(soundChan)
			close(visualizeChan)
			close(addWordChan)
			close(resultsChan)
		}()

		for result := range l.translateWords(ctx) {
			if result.Error != nil {
				slog.Error("cannot translate word", "word", result.Result.Word, "error", result.Error)
				continue
			}
			if l.Sound {
				soundChan <- result.Result.SoundURL
			}
			if l.Visualise {
				for _, p := range result.Result.Translate {
					if p.Picture == "" {
						continue
					}
					url, err := url.Parse(p.Picture)
					if err != nil {
						slog.Error("cannot parse picture url", "url", p.Picture, "error", err)
						continue
					}
					visualizeChan <- url
				}
			}

			if l.Add {
				if resultsToAdd := l.prepareResultToAdd(&result.Result); resultsToAdd {
					addWordChan <- result.Result
				}
			}
			resultsChan <- result.Result
		}
	}()

	return Channels{
		sound:     soundChan,
		visualize: visualizeChan,
		add:       addWordChan,
		results:   resultsChan,
	}
}

type processResult func(api.Result) error

func ProcessResultImpl(r api.Result) error {
	r.PrintTranslation()
	return nil
}

func (l *Lingualeo) translateToChan(ctx context.Context) <-chan api.Result {
	var wg sync.WaitGroup
	wg.Add(1)
	channels := l.Process(ctx, &wg)
	if l.Sound {
		wg.Add(1)
		go l.Pronounce(ctx, channels.sound, &wg)
	}
	if l.Add {
		wg.Add(1)
		go l.AddToDictionary(ctx, channels.add, &wg)
	}
	if l.Visualise {
		wg.Add(1)
		go l.Visualize(ctx, channels.visualize, &wg)
	}

	ch := make(chan api.Result, len(l.Words))

	go func() {
		defer close(ch)
		for result := range channel.OrDone(ctx, channels.results) {
			ch <- result
		}
		wg.Wait()
	}()

	return ch
}

func (l *Lingualeo) TranslateWithReverseRussian(ctx context.Context, resultFunc processResult) {
	// TranslateWithReverseRussian translates russian words,
	// gets english translations and translates them once more
	var englishWords []string
	for result := range channel.OrDone(ctx, l.translateToChan(ctx)) {
		if err := resultFunc(result); err != nil {
			slog.Error("cannot translate word", "word", result.Word, "error", err)
		}
		for _, word := range result.Translate {
			if l.ReverseTranslate && isEnglishWord(word.Value) {
				englishWords = append(englishWords, word.Value)
			}
		}
	}
	if len(englishWords) > 0 {
		l.Words = englishWords
		for result := range channel.OrDone(ctx, l.translateToChan(ctx)) {
			if err := resultFunc(result); err != nil {
				slog.Error("cannot translate word", "word", result.Word, "error", err)
			}
		}
	}
}
