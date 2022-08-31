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

// Result API result interface
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

// Word translates word structure
type Word struct {
	ID      int                `json:"id"`
	Votes   int                `json:"votes"`
	Value   string             `json:"value"`
	Picture string             `json:"pic_url"`
	Exists  convertibleBoolean `json:"ut"`
}

// OpResult represents operation result
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

// TranslationResult represents API response
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

// FromResponse fills TranslationResult from http response
func (result *TranslationResult) FromResponse(body string) error {
	return fromResponse(result, body)
}

// GetTranscription returns word transcription
func (result *TranslationResult) GetTranscription() []string {
	return []string{result.Transcription}
}

// PrintTranslation prints transcription
func (result *TranslationResult) PrintTranslation() {
	printTranslation(result)
}

// PrintAddedTranslation prints transcription during adding operation
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

// GetWord returns word to translate
func (result *TranslationResult) GetWord() string {
	return result.Word
}

// SetWord sets word to translate
func (result *TranslationResult) SetWord(word string) {
	result.Word = word
}

// GetTranslate returns translation for a word
func (result *TranslationResult) GetTranslate() []string {
	return result.Words
}

// SetTranslate sets custom translation for a word
func (result *TranslationResult) SetTranslate(translates []string, replace bool) {
	if replace {
		result.Words = utils.Unique(translates)
	} else {
		result.Words = utils.Unique(append(result.Words, translates...))
	}
}

// GetSoundURLs returns sound urls to pronounce
func (result *TranslationResult) GetSoundURLs() []string {
	return []string{result.SoundURL}
}

// InDictionary checks either word is already has been added into the dictionary
func (result *TranslationResult) InDictionary() bool {
	if bool(result.Exists) {
		return true
	}
	for _, word := range result.Translate {
		if bool(word.Exists) {
			return true
		}
	}
	return false
}

func (result *TranslationResult) Error() string {
	return result.ErrorMsg
}

// IsRussian either word in in Russian language
func (result *TranslationResult) IsRussian() bool {
	return result.Transcription == ""
}

// NoResult negative operation result
type NoResult struct {
	Translate []string `json:"translate"`
	ErrorMsg  string   `json:"error_msg"`
}
