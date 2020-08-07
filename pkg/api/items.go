package api

import (
	"fmt"
	"sort"
	"strings"

	"github.com/trezorg/lingualeo/pkg/utils"
)

type convertibleBoolean bool

func (bit *convertibleBoolean) UnmarshalJSON(data []byte) error {
	asString := strings.Trim(string(data), "\"")
	if asString == "1" || asString == "true" {
		*bit = true
	} else if asString == "0" || asString == "false" || asString == "null" {
		*bit = false
	} else {
		return fmt.Errorf("boolean unmarshal error: invalid input %s", asString)
	}
	return nil
}

type Result interface {
	GetWord() string
	Error() string
	SetWord(string)
	GetTranslate() []string
	GetTranscription() []string
	SetTranslate([]string, bool)
	GetSoundURLs() []string
	FromResponse(body string) error
	parse()
	PrintTranslation()
	PrintAddedTranslation()
	InDictionary() bool
	IsRussian() bool
}

type apiError struct {
	ErrorMsg  string `json:"error_msg"`
	ErrorCode int    `json:"error_code"`
}

type Word struct {
	ID      int                `json:"id"`
	Votes   int                `json:"votes"`
	Value   string             `json:"value"`
	Picture string             `json:"pic_url"`
	Exists  convertibleBoolean `json:"is_user"`
}

type OpResult struct {
	Error  error
	Result Result
}

func opResultFromBody(word string, body string) OpResult {
	res := TranslationResult{Word: word}
	err := res.FromResponse(body)
	return OpResult{
		Error:  err,
		Result: &res,
	}
}

type TranslationResult struct {
	Word             string             `json:"-"`
	Words            []string           `json:"-"`
	Exists           convertibleBoolean `json:"is_user"`
	SoundURL         string             `json:"sound_url"`
	Transcription    string             `json:"transcription"`
	Translate        []Word             `json:"translate"`
	ErrorMsg         string             `json:"error_msg"`
	DirectionEnglish bool               `json:"directionEnglish"`
}

func (result *TranslationResult) FromResponse(body string) error {
	return fromResponse(result, body)
}

func (result *TranslationResult) GetTranscription() []string {
	return []string{result.Transcription}
}

func (result *TranslationResult) PrintTranslation() {
	printTranslation(result)
}

func (result *TranslationResult) PrintAddedTranslation() {
	printAddedTranslation(result)
}

func (result *TranslationResult) parse() {
	sort.Slice(result.Translate, func(i, j int) bool {
		return result.Translate[i].Votes > result.Translate[j].Votes
	})
	for _, translate := range result.Translate {
		result.Words = append(result.Words, translate.Value)
	}
	result.Words = utils.Unique(result.Words)
}

func (result *TranslationResult) GetWord() string {
	return result.Word
}

func (result *TranslationResult) SetWord(word string) {
	result.Word = word
}

func (result *TranslationResult) GetTranslate() []string {
	return result.Words
}

func (result *TranslationResult) SetTranslate(translates []string, replace bool) {
	if replace {
		result.Words = utils.Unique(translates)
	} else {
		result.Words = utils.Unique(append(result.Words, translates...))
	}
}

func (result *TranslationResult) GetSoundURLs() []string {
	return []string{result.SoundURL}
}

func (result *TranslationResult) InDictionary() bool {
	return bool(result.Exists)
}

func (result *TranslationResult) Error() string {
	return result.ErrorMsg
}

func (result *TranslationResult) IsRussian() bool {
	return result.Transcription == ""
}

type NoResult struct {
	Translate []string `json:"translate"`
	ErrorMsg  string   `json:"error_msg"`
}
