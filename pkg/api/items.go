package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/trezorg/lingualeo/internal/slice"
	"github.com/trezorg/lingualeo/pkg/messages"
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
	Value     string             `json:"value"`
	Picture   string             `json:"pic_url"`
	Context   string             `json:"ctx"`
	Translate string             `json:"tr"`
	ID        int                `json:"id"`
	Votes     int                `json:"votes"`
	Exists    convertibleBoolean `json:"ut"`
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

func readBody(resp *http.Response) ([]byte, error) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Error("cannot close response body", "error", err)
		}
	}()
	return io.ReadAll(resp.Body)
}

// Result represents API response
type Result struct {
	Word                     string             `json:"-"`
	SoundURL                 string             `json:"sound_url"`
	Transcription            string             `json:"transcription"`
	ErrorMsg                 string             `json:"error_msg"`
	Pos                      string             `json:"pos"`
	AddWords                 []string           `json:"-"`
	Translate                []Word             `json:"translate"`
	Exists                   convertibleBoolean `json:"is_user"`
	DirectionEnglish         bool               `json:"directionEnglish"`
	InvertTranslateDirection bool               `json:"invertTranslateDirection"`
}

// FromResponse fills TranslationResult from http response
func (result *Result) FromResponse(body []byte) error {
	err := json.Unmarshal(body, &result)
	if err != nil {
		res := NoResult{}
		if fErr := json.Unmarshal(body, &res); fErr != nil {
			return fmt.Errorf("cannot translate word: %s, %w", result.Word, fErr)
		}
		return err
	}
	if len(result.Error()) > 0 {
		return result
	}
	result.parse()
	return nil
}

// PrintTranslation prints transcription
func (result *Result) PrintTranslation() {
	var strTitle string
	if result.InDictionary() {
		strTitle = "existing"
	} else {
		strTitle = "new"
	}
	if err := messages.Message(messages.RED, "Found %s word:\n", strTitle); err != nil {
		slog.Error("cannot show message", "error", err)
	}
	if err := messages.Message(messages.GREEN, "['%s'] (%s)\n", result.Word, result.Transcription); err != nil {
		slog.Error("cannot show message", "error", err)
	}
	for _, word := range result.Translate {
		if err := messages.Message(messages.YELLOW, "%s", word.Value); err != nil {
			slog.Error("cannot show message", "error", err)
		}
		if len(word.Context) > 0 {
			if err := messages.Message(messages.WHITE, " (%s)", word.Context); err != nil {
				slog.Error("cannot show message", "error", err)
			}
		}
		if err := messages.Message(messages.YELLOW, "\n"); err != nil {
			slog.Error("cannot show message", "error", err)
		}
	}
}

// PrintAddedTranslation prints transcription during adding operation
func (result *Result) PrintAddedTranslation() {
	var strTitle string
	if result.InDictionary() {
		strTitle = "Updated existing"
	} else {
		strTitle = "Added new"
	}
	err := messages.Message(messages.RED, "%s word: ", strTitle)
	if err != nil {
		slog.Error("cannot show message", "error", err)
	}

	err = messages.Message(messages.GREEN, "['%s'] ['%s']\n", result.Word, strings.Join(result.AddWords, ", "))
	if err != nil {
		slog.Error("cannot show message", "error", err)
	}
}

func (result *Result) parse() {
	result.Translate = slice.UniqueFunc(result.Translate, func(w Word) string { return w.Value })
	sort.Slice(result.Translate, func(i, j int) bool {
		return result.Translate[i].Votes > result.Translate[j].Votes
	})
}

// SetTranslation sets custom translation for a word
func (result *Result) SetTranslation(translates []string) {
	result.AddWords = slice.Unique(translates)
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
