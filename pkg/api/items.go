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
	engOp := opEnglishResultFromBody(word, body)
	if engOp.Error != nil {
		rusOp := opRussianResultFromBody(word, body)
		if rusOp.Error != nil {
			return engOp
		}
		return rusOp
	}
	return engOp
}

func opEnglishResultFromBody(word string, body string) OpResult {
	res := EnglishResult{Word: word}
	err := res.FromResponse(body)
	return OpResult{
		Error:  err,
		Result: &res,
	}
}

func opRussianResultFromBody(word string, body string) OpResult {
	res := RussianResult{Word: word}
	err := res.FromResponse(body)
	return OpResult{
		Error:  err,
		Result: &res,
	}
}

type WordForm struct {
	PictureURL    string `json:"pic_url"`
	SoundURL      string `json:"sound_url"`
	Transcription string `json:"transcription"`
	Votes         int    `json:"votes"`
	Word          string `json:"word"`
}

type EnglishResult struct {
	Word             string   `json:"-"`
	Words            []string `json:"-"`
	Exists           bool     `json:"-"`
	SoundURL         string   `json:"sound_url"`
	Transcription    string   `json:"transcription"`
	Translate        []Word   `json:"translate"`
	ErrorMsg         string   `json:"error_msg"`
	DirectionEnglish bool     `json:"directionEnglish"`
}

func (result *EnglishResult) FromResponse(body string) error {
	return fromResponse(result, body)
}

func (result *EnglishResult) GetTranscription() []string {
	return []string{result.Transcription}
}

func (result *EnglishResult) PrintTranslation() {
	printTranslation(result)
}

func (result *EnglishResult) PrintAddedTranslation() {
	printAddedTranslation(result)
}

func (result *EnglishResult) parse() {
	isUsed := false
	sort.Slice(result.Translate, func(i, j int) bool {
		return result.Translate[i].Votes > result.Translate[j].Votes
	})
	for _, translate := range result.Translate {
		if !isUsed {
			isUsed = bool(translate.Exists)
		}
		result.Words = append(result.Words, sanitizeWords(translate.Value)...)
	}
	result.Exists = isUsed
	result.Words = utils.Unique(result.Words)
}

func (result *EnglishResult) GetWord() string {
	return result.Word
}

func (result *EnglishResult) SetWord(word string) {
	result.Word = word
}

func (result *EnglishResult) GetTranslate() []string {
	return result.Words
}

func (result *EnglishResult) SetTranslate(translates []string, replace bool) {
	if replace {
		result.Words = utils.Unique(translates)
	} else {
		result.Words = utils.Unique(append(result.Words, translates...))
	}
}

func (result *EnglishResult) GetSoundURLs() []string {
	return []string{result.SoundURL}
}

func (result *EnglishResult) InDictionary() bool {
	return result.Exists
}

func (result *EnglishResult) Error() string {
	return result.ErrorMsg
}

type NoResult struct {
	Translate []string `json:"translate"`
	ErrorMsg  string   `json:"error_msg"`
}

type RussianResult struct {
	Word             string             `json:"-"`
	Words            []string           `json:"-"`
	Exists           convertibleBoolean `json:"is_user"`
	Translate        []string           `json:"translate"`
	ErrorMsg         string             `json:"error_msg"`
	DirectionEnglish bool               `json:"directionEnglish"`
	WordForms        []WordForm         `json:"word_forms"`
}

func (result *RussianResult) FromResponse(body string) error {
	return fromResponse(result, body)
}

func (result *RussianResult) parse() {
	sort.Slice(result.WordForms, func(i, j int) bool {
		return result.WordForms[i].Votes > result.WordForms[j].Votes
	})
	for _, form := range result.WordForms {
		result.Words = append(result.Words, form.Word)
	}
	result.Words = utils.Unique(result.Words)
}

func (result *RussianResult) Error() string {
	return result.ErrorMsg
}

func (result *RussianResult) InDictionary() bool {
	return bool(result.Exists)
}

func (result *RussianResult) GetSoundURLs() []string {
	urls := make([]string, 0)
	for _, form := range result.WordForms {
		urls = append(urls, form.SoundURL)
	}
	return urls
}

func (result *RussianResult) GetWord() string {
	return result.Word
}

func (result *RussianResult) SetWord(word string) {
	result.Word = word
}

func (result *RussianResult) GetTranslate() []string {
	return result.Words
}

func (result *RussianResult) GetTranscription() []string {
	transcriptions := make([]string, 0)
	for _, form := range result.WordForms {
		transcriptions = append(transcriptions, form.Transcription)
	}
	return transcriptions
}

func (result *RussianResult) SetTranslate(translates []string, replace bool) {
	if replace {
		result.Words = utils.Unique(translates)
	} else {
		result.Words = utils.Unique(append(result.Words, translates...))
	}
}

func (result *RussianResult) PrintTranslation() {
	printTranslation(result)
}

func (result *RussianResult) PrintAddedTranslation() {
	printAddedTranslation(result)
}
