package api

import (
	"fmt"
	"sort"
	"strings"

	"github.com/trezorg/lingualeo/internal/slice"
)

type convertibleBoolean bool

func (bit *convertibleBoolean) UnmarshalJSON(data []byte) error {
	asString := strings.Trim(string(data), "\"")

	switch asString {
	case "1", "true":
		*bit = true
	case "0", "false", "null":
		*bit = false
	default:
		return fmt.Errorf("boolean unmarshal error: invalid input %s", asString)
	}

	return nil
}

type apiError struct {
	ErrorMsg  string `json:"error_msg"`
	ErrorCode int    `json:"error_code"`
}

// Word translates word structure
type Word struct {
	Value   string             `json:"value"`
	Picture string             `json:"pic_url"`
	ID      int                `json:"id"`
	Votes   int                `json:"votes"`
	Exists  convertibleBoolean `json:"ut"`
}

// OperationResult represents operation result
type OperationResult struct {
	Error  error
	Result Result
}

func opResultFromBody(word string, body []byte) OperationResult {
	res := Result{Word: word}
	err := res.FromResponse(body)
	return OperationResult{
		Error:  err,
		Result: res,
	}
}

// Result represents API response
type Result struct {
	Word             string             `json:"-"`
	SoundURL         string             `json:"sound_url"`
	Transcription    string             `json:"transcription"`
	ErrorMsg         string             `json:"error_msg"`
	Words            []string           `json:"-"`
	Translate        []Word             `json:"translate"`
	Exists           convertibleBoolean `json:"is_user"`
	DirectionEnglish bool               `json:"directionEnglish"`
}

// FromResponse fills TranslationResult from http response
func (result *Result) FromResponse(body []byte) error {
	return fromResponse(result, body)
}

// PrintTranslation prints transcription
func (result *Result) PrintTranslation() {
	printTranslation(result)
}

// PrintAddedTranslation prints transcription during adding operation
func (result *Result) PrintAddedTranslation() {
	printAddedTranslation(result)
}

func (result *Result) parse() {
	sort.Slice(result.Translate, func(i, j int) bool {
		return result.Translate[i].Votes > result.Translate[j].Votes
	})
	for _, translate := range result.Translate {
		result.Words = append(result.Words, translate.Value)
	}
	result.Words = slice.Unique(result.Words)
}

// SetTranslate sets custom translation for a word
func (result *Result) SetTranslate(translates []string, replace bool) {
	if replace {
		result.Words = slice.Unique(translates)
	} else {
		result.Words = slice.Unique(append(result.Words, translates...))
	}
}

// InDictionary checks either word is already has been added into the dictionary
func (result *Result) InDictionary() bool {
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

func (result *Result) Error() string {
	return result.ErrorMsg
}

// IsRussian either word in in Russian language
func (result *Result) IsRussian() bool {
	return result.Transcription == ""
}

// NoResult negative operation result
type NoResult struct {
	ErrorMsg  string   `json:"error_msg"`
	Translate []string `json:"translate"`
}
