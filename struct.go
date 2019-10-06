package main

import (
	"io"
	"sort"
)

type lingualeoWordResult struct {
	Id      int                `json:"id"`
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
		for _, word := range sanitizeWords(&translate) {
			result.Words = append(result.Words, word)
		}
	}
	result.Exists = isUsed
	result.Words = unique(result.Words)
}

func (result *lingualeoResult) findWordUsage() {
	sort.Slice(result.Translate, func(i, j int) bool {
		return result.Translate[i].Votes > result.Translate[j].Votes
	})
	for _, translate := range result.Translate {
		for _, word := range sanitizeWords(&translate) {
			result.Words = append(result.Words, word)
		}
	}
	result.Words = unique(result.Words)
}

func (result *lingualeoResult) fillObjectFromJSON(body io.ReadCloser) error {
	return getJSON(body, result)
}

type lingualeoArgs struct {
	Email          string `yaml:"email"`
	Password       string `yaml:"password"`
	Config         string
	Player         string `yaml:"player"`
	Words          []string
	Translate      []string
	Force          bool   `yaml:"force"`
	Add            bool   `yaml:"addTranslationToDictionary"`
	Sound          bool   `yaml:"sound"`
	LogLevel       string `yaml:"log_level"`
	LogPrettyPrint bool   `yaml:"log_pretty_print"`
}
