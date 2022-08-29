package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/trezorg/lingualeo/pkg/messages"

	"github.com/trezorg/lingualeo/pkg/logger"
)

func getJSONFromString(body string, target interface{}) error {
	return json.Unmarshal([]byte(body), &target)
}

func readBody(resp *http.Response) (*string, error) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Error(err)
		}
	}()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := string(data)
	return &res, nil
}

func fromResponse(result Result, body string) error {
	err := getJSONFromString(body, result)
	if err != nil {
		res := NoResult{}
		if err := getJSONFromString(body, res); err != nil {
			return fmt.Errorf("cannot translate word: %s, %w", result.GetWord(), err)
		}
		return err
	}
	if len(result.Error()) > 0 {
		return result
	}
	result.parse()
	return nil
}

func printTranslation(result Result) {
	var strTitle string
	if result.InDictionary() {
		strTitle = "existing"
	} else {
		strTitle = "new"
	}
	err := messages.Message(messages.RED, "Found %s word:\n", strTitle)
	if err != nil {
		logger.Error(err)
	}
	err = messages.Message(messages.GREEN, "['%s'] (%s)\n", result.GetWord(), strings.Join(result.GetTranscription(), ", "))
	if err != nil {
		logger.Error(err)
	}
	for _, word := range result.GetTranslate() {
		_ = messages.Message(messages.YELLOW, "%s\n", word)
	}
}

func printAddedTranslation(result Result) {
	var strTitle string
	if result.InDictionary() {
		strTitle = "Updated existing"
	} else {
		strTitle = "Added new"
	}
	err := messages.Message(messages.RED, "%s word: ", strTitle)
	if err != nil {
		logger.Error(err)
	}
	err = messages.Message(messages.GREEN, "['%s']\n", result.GetWord())
	if err != nil {
		logger.Error(err)
	}
}
