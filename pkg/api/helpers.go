package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/trezorg/lingualeo/pkg/messages"
)

func readBody(resp *http.Response) ([]byte, error) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Error("cannot close response body", "error", err)
		}
	}()
	return io.ReadAll(resp.Body)
}

func fromResponse(result *Result, body []byte) error {
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

func printTranslation(result *Result) {
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

func printAddedTranslation(result *Result) {
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
