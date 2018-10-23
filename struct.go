package main

import (
	"io"
	"sort"
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

func (result *lingualeoResult) parseAndSortTranslate() {
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
	Email     string
	Password  string
	Config    string
	Player    string
	Words     []string
	Translate []string
	Force     bool
	Add       bool
	Sound     bool
}
