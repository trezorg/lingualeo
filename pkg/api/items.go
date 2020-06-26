package api

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/trezorg/lingualeo/pkg/logger"
	"github.com/trezorg/lingualeo/pkg/utils"

	"github.com/wsxiaoys/terminal/color"
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

type Result struct {
	Word          string   `json:"-"`
	Words         []string `json:"-"`
	Exists        bool     `json:"-"`
	SoundURL      string   `json:"sound_url"`
	Transcription string   `json:"transcription"`
	Translate     []Word   `json:"translate"`
	ErrorMsg      string   `json:"error_msg"`
}

func (result *Result) FromJSON(body io.ReadCloser) error {
	return utils.GetJSON(body, result)
}

func (result *Result) PrintTranslation() {
	var strTitle string
	if result.Exists {
		strTitle = "existing"
	} else {
		strTitle = "new"
	}
	_, err := color.Printf("@{r}Found %s word:\n", strTitle)
	if err != nil {
		logger.Log.Error(err)
	}
	_, err = color.Printf("@{g}['%s'] (%s)\n", result.Word, result.Transcription)
	if err != nil {
		logger.Log.Error(err)
	}
	for _, word := range result.Words {
		utils.PrintColorString("y", word)
	}
}

func (result *Result) PrintAddedTranslation() {
	var strTitle string
	if result.Exists {
		strTitle = "Updated existing"
	} else {
		strTitle = "Added new"
	}
	_, err := color.Printf("@{r}%s word: ", strTitle)
	if err != nil {
		logger.Log.Error(err)
	}
	_, err = color.Printf("@{g}['%s']\n", result.Word)
	if err != nil {
		logger.Log.Error(err)
	}
}

func (result *Result) ParseTranslation() {
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

type NoResult struct {
	Translate []string `json:"translate"`
	ErrorMsg  string   `json:"error_msg"`
}
