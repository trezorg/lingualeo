package translator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/trezorg/lingualeo/internal/api"
	"github.com/trezorg/lingualeo/internal/channel"
	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/internal/messages"
	"github.com/trezorg/lingualeo/internal/visualizer/browser"
	"github.com/trezorg/lingualeo/internal/visualizer/term"

	"github.com/trezorg/lingualeo/internal/player"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var errUnknownVisualiseType = errors.New("unknown visualize type")

type Lingualeo struct {
	api.Client            `json:"-" yaml:"-" toml:"-"`
	Downloader            `json:"-" yaml:"-" toml:"-"`
	Pronouncer            `json:"-" yaml:"-" toml:"-"`
	Outputer              `json:"-" yaml:"-" toml:"-"`
	Email                 string        `yaml:"email" json:"email" toml:"email"`
	VisualiseType         VisualiseType `yaml:"visualize_type" json:"visualize_type" toml:"visualize_type"`
	Config                string
	Player                string        `yaml:"player" json:"player" toml:"player"`
	LogLevel              string        `yaml:"log_level" json:"log_level" toml:"log_level"`
	Password              string        `yaml:"password" json:"password" toml:"password"` //nolint:gosec // credential field
	RequestTimeout        time.Duration `yaml:"request_timeout" json:"request_timeout" toml:"request_timeout"`
	PlayerShutdownTimeout time.Duration `yaml:"player_shutdown_timeout" json:"player_shutdown_timeout" toml:"player_shutdown_timeout"`
	// HTTP connection pool settings
	MaxIdleConns        int           `yaml:"max_idle_conns" json:"max_idle_conns" toml:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" json:"max_idle_conns_per_host" toml:"max_idle_conns_per_host"`
	MaxRedirects        int           `yaml:"max_redirects" json:"max_redirects" toml:"max_redirects"`
	RetryMaxAttempts    int           `yaml:"retry_max_attempts" json:"retry_max_attempts" toml:"retry_max_attempts"`
	RetryInitialWait    time.Duration `yaml:"retry_initial_wait" json:"retry_initial_wait" toml:"retry_initial_wait"`
	RetryMaxWait        time.Duration `yaml:"retry_max_wait" json:"retry_max_wait" toml:"retry_max_wait"`
	Translation         []string
	Words               []string
	Add                 bool `yaml:"add" json:"add" toml:"add"`
	Sound               bool `yaml:"sound" json:"sound" toml:"sound"`
	Visualise           bool `yaml:"visualize" json:"visualize" toml:"visualize"`
	Debug               bool `yaml:"debug" json:"debug" toml:"debug"`
	DownloadSoundFile   bool `yaml:"download" json:"download" toml:"download"`
	LogPrettyPrint      bool `yaml:"log_pretty_print" json:"log_pretty_print" toml:"log_pretty_print"`
	ReverseTranslate    bool `yaml:"reverse_translate" json:"reverse_translate" toml:"reverse_translate"`
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

// sendOperationResult sends a result to the output channel.
// The return value is intentionally ignored as this is a fire-and-forget
// operation - if the context is cancelled, the result is simply dropped.
// sendOperationResult sends a result to the output channel.
// The return value is intentionally ignored as this is a fire-and-forget
// operation - if the context is cancelled, the result is simply dropped.
// This is by design: when the pipeline is shutting down, we don't want to
// block on sending results that won't be processed anyway.
func sendOperationResult(ctx context.Context, out chan<- api.OperationResult, res api.OperationResult) {
	_ = sendWithContext(ctx, out, res)
}

func sendWithContext[T any](ctx context.Context, out chan<- T, value T) bool {
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
	if err = client.checkMediaPlayer(); err != nil {
		slog.Warn("media player check failed", "error", err)
	}

	for _, option := range options {
		if err = option(&client); err != nil {
			return client, err
		}
	}
	if client.Client == nil {
		timeout := client.RequestTimeout
		if timeout == 0 {
			timeout = defaultRequestTimeout
		}
		apiCfg := api.Config{
			Timeout:             timeout,
			MaxRedirects:        client.MaxRedirects,
			MaxIdleConns:        client.MaxIdleConns,
			MaxIdleConnsPerHost: client.MaxIdleConnsPerHost,
			Retry: api.RetryConfig{
				MaxAttempts: client.RetryMaxAttempts,
				InitialWait: client.RetryInitialWait,
				MaxWait:     client.RetryMaxWait,
			},
		}
		if client.Client, err = api.New(context.Background(), client.Email, client.Password, client.Debug, apiCfg); err != nil {
			return client, err
		}
	}
	if client.Pronouncer == nil {
		opts := make([]player.Option, 0, 1)
		if client.PlayerShutdownTimeout > 0 {
			opts = append(opts, player.WithShutdownTimeout(client.PlayerShutdownTimeout))
		}
		client.Pronouncer = player.New(client.Player, opts...)
	}
	if client.Downloader == nil {
		client.Downloader = files.New()
	}
	if client.Outputer == nil {
		client.Outputer, err = outputer(client.Visualise, client.VisualiseType)
	}
	return client, err
}

// translateWords translate words from string channel
func translateWords(ctx context.Context, translator api.Client, results <-chan string) <-chan api.OperationResult {
	out := make(chan api.OperationResult)
	var wg sync.WaitGroup
	wg.Go(func() {
		for word := range channel.OrDone(ctx, results) {
			wg.Go(func() {
				sendOperationResult(ctx, out, translator.TranslateWord(ctx, word))
			})
		}
	})
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// addWords add words
func addWords(ctx context.Context, translator api.Client, results <-chan api.Result) <-chan api.OperationResult {
	out := make(chan api.OperationResult)
	var wg sync.WaitGroup
	wg.Go(func() {
		for res := range channel.OrDone(ctx, results) {
			for _, translate := range res.AddWords {
				translation := translate
				wg.Go(func() {
					added := translator.AddWord(ctx, res.Word, translation)
					added.Result.AddWords = []string{translation}
					sendOperationResult(ctx, out, added)
				})
			}
		}
	})
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

func (l *Lingualeo) translateWords(ctx context.Context) <-chan api.OperationResult {
	results := make(chan api.OperationResult, len(l.Words))
	input := channel.ToChannel(ctx, l.Words...)
	go func() {
		defer close(results)
		ch := translateWords(ctx, l.Client, input)
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
			if !sendWithContext(ctx, results, res) {
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
	return false
}

func (l *Lingualeo) downloadAndPronounce(ctx context.Context, urls <-chan string) {
	fileChannel := files.OrderedChannel(downloadFiles(ctx, urls, l.Downloader), len(l.Words))
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
func (l *Lingualeo) Pronounce(ctx context.Context, urls <-chan string) {
	if l.DownloadSoundFile {
		l.downloadAndPronounce(ctx, urls)
		return
	}
	l.playURLs(ctx, urls)
}

// AddToDictionary adds words to dictionary
func (l *Lingualeo) AddToDictionary(ctx context.Context, resultsToAdd <-chan api.Result) {
	ch := addWords(ctx, l.Client, resultsToAdd)
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

// Process starts translation process
func (l *Lingualeo) Process(ctx context.Context, wg *sync.WaitGroup) Channels {
	soundChan := make(chan string, len(l.Words))
	addWordChan := make(chan api.Result, len(l.Translation))
	resultsChan := make(chan api.Result, len(l.Words))

	go func() {
		defer func() {
			wg.Done()
			close(soundChan)
			close(addWordChan)
			close(resultsChan)
		}()

		for result := range l.translateWords(ctx) {
			if result.Error != nil {
				slog.Error("cannot translate word", "word", result.Result.Word, "error", result.Error)
				continue
			}
			if l.Sound {
				if !sendWithContext(ctx, soundChan, result.Result.SoundURL) {
					return
				}
			}

			if l.Add {
				if resultsToAdd := l.prepareResultToAdd(&result.Result); resultsToAdd {
					if !sendWithContext(ctx, addWordChan, result.Result) {
						return
					}
				}
			}
			if !sendWithContext(ctx, resultsChan, result.Result) {
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

func (l *Lingualeo) translateToChan(ctx context.Context) <-chan api.Result {
	var wg sync.WaitGroup
	wg.Add(1)
	channels := l.Process(ctx, &wg)
	if l.Sound {
		wg.Go(func() {
			l.Pronounce(ctx, channels.sound)
		})
	}
	if l.Add {
		wg.Go(func() {
			l.AddToDictionary(ctx, channels.add)
		})
	}

	ch := make(chan api.Result, len(l.Words))

	go func() {
		defer close(ch)
		for result := range channel.OrDone(ctx, channels.results) {
			if !sendWithContext(ctx, ch, result) {
				return
			}
		}
		wg.Wait()
	}()

	return ch
}

func (l *Lingualeo) TranslateWithReverseRussian(ctx context.Context) {
	// TranslateWithReverseRussian translates russian words,
	// gets english translations and translates them once more
	var englishWords []string
	for result := range channel.OrDone(ctx, l.translateToChan(ctx)) {
		if err := l.Output(ctx, result); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
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
			if err := l.Output(ctx, result); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				slog.Error("cannot translate word", "word", result.Word, "error", err)
			}
		}
	}
}
