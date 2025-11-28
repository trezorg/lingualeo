package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/trezorg/lingualeo/internal/slice"
)

type convertibleBoolean bool

var (
	errBooleanUnmarshal = errors.New("boolean unmarshal error")
	errTranslateWord    = errors.New("cannot translate word")
)

func (bit *convertibleBoolean) UnmarshalJSON(data []byte) error {
	asString := strings.Trim(string(data), "\"")

	switch asString {
	case "1", "true":
		*bit = true
	case "0", "false", "null":
		*bit = false
	default:
		return fmt.Errorf("%w: invalid input %s", errBooleanUnmarshal, asString)
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

// ResultError wraps translation results that failed on the API side.
type ResultError struct {
	Result Result
}

func (e ResultError) Error() string {
	if len(e.Result.ErrorMsg) == 0 {
		return errTranslateWord.Error()
	}
	if len(e.Result.Word) > 0 {
		return fmt.Sprintf("%s: %s", e.Result.Word, e.Result.ErrorMsg)
	}
	return e.Result.ErrorMsg
}

func (ResultError) Unwrap() error {
	return errTranslateWord
}

// FromResponse fills TranslationResult from http response
func (r *Result) FromResponse(body []byte) error {
	err := json.Unmarshal(body, &r)
	if err != nil {
		res := NoResult{}
		if fErr := json.Unmarshal(body, &res); fErr != nil {
			return fmt.Errorf("%w: %s, %w", errTranslateWord, r.Word, fErr)
		}
		return err
	}
	if r.HasError() {
		return ResultError{Result: *r}
	}
	r.parse()
	return nil
}

func (r *Result) parse() {
	r.Translate = slice.UniqueFunc(r.Translate, func(w Word) string { return w.Value })
	sort.Slice(r.Translate, func(i, j int) bool {
		return r.Translate[i].Votes > r.Translate[j].Votes
	})
}

// SetTranslation sets custom translation for a word
func (r *Result) SetTranslation(translates []string) {
	r.AddWords = slice.Unique(translates)
}

// InDictionary checks either word is already has been added into the dictionary
func (r *Result) InDictionary() bool {
	if bool(r.Exists) {
		return true
	}
	for _, word := range r.Translate {
		if bool(word.Exists) {
			return true
		}
	}
	return false
}

func (r *Result) HasError() bool {
	return len(r.ErrorMsg) > 0
}

// IsRussian either word in in Russian language
func (r *Result) IsRussian() bool {
	return r.Transcription == ""
}

// NoResult negative operation result
type NoResult struct {
	ErrorMsg  string   `json:"error_msg"`
	Translate []string `json:"translate"`
}
