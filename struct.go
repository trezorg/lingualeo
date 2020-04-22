package main

import (
	"io"
	"sort"

	"github.com/urfave/cli/v2"
)

type indexedItem interface {
	getIndex() int
}

type lingualeoWordResult struct {
	ID      int                `json:"id"`
	Votes   int                `json:"votes"`
	Value   string             `json:"value"`
	Picture string             `json:"pic_url"`
	Exists  convertibleBoolean `json:"is_user"`
}

type responseError struct {
	ErrorMsg  string `json:"error_msg"`
	ErrorCode int    `json:"error_code"`
}

type translateResult struct {
	Error  error
	Result lingualeoResult
}

type resultFile struct {
	Error    error
	Filename string
	Index    int
}

func (rf resultFile) getIndex() int {
	return rf.Index
}

type lingualeoResult struct {
	Word          string                `json:"-"`
	Words         []string              `json:"-"`
	Exists        bool                  `json:"-"`
	SoundURL      string                `json:"sound_url"`
	Transcription string                `json:"transcription"`
	Translate     []lingualeoWordResult `json:"translate"`
	ErrorMsg      string                `json:"error_msg"`
}

type lingualeoNoResult struct {
	Translate []string `json:"translate"`
	ErrorMsg  string   `json:"error_msg"`
}

func (result *lingualeoResult) parseAndSortTranslate() {
	isUsed := false
	sort.Slice(result.Translate, func(i, j int) bool {
		return result.Translate[i].Votes > result.Translate[j].Votes
	})
	for _, translate := range result.Translate {
		if !isUsed {
			isUsed = bool(translate.Exists)
		}
		result.Words = append(result.Words, sanitizeWords(&translate)...)
	}
	result.Exists = isUsed
	result.Words = unique(result.Words)
}

func (result *lingualeoResult) fillObjectFromJSON(body io.ReadCloser) error {
	return getJSON(body, result)
}

type lingualeoArgs struct {
	Email             string `yaml:"email" json:"email" toml:"email"`
	Password          string `yaml:"password" json:"password" toml:"password"`
	Config            string
	Player            string `yaml:"player" json:"player" toml:"player"`
	Words             []string
	Translate         cli.StringSlice
	Force             bool   `yaml:"force" json:"force" toml:"force"`
	Add               bool   `yaml:"add" json:"add" toml:"add"`
	TranslateReplace  bool   `yaml:"translate_replace" json:"translate_replace" toml:"translate_replace"`
	Sound             bool   `yaml:"sound" json:"sound" toml:"sound"`
	Debug             bool   `yaml:"debug" json:"debug" toml:"debug"`
	DownloadSoundFile bool   `yaml:"download" json:"download" toml:"download"`
	LogLevel          string `yaml:"log_level" json:"log_level" toml:"log_level"`
	LogPrettyPrint    bool   `yaml:"log_pretty_print" json:"log_pretty_print" toml:"log_pretty_print"`
}
