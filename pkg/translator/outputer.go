package translator

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	api "github.com/trezorg/lingualeo/pkg/api"
	"github.com/trezorg/lingualeo/pkg/messages"
)

type Outputer interface {
	Output(ctx context.Context, r api.Result) error
}

func parseURL(s string) (*url.URL, error) {
	if s == "" {
		return nil, nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("cannot parse picture url: %s", s)
	}
	return u, nil
}

func (o Output) Output(ctx context.Context, result api.Result) error {
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
		select {
		case <-ctx.Done():
			return nil
		default:
			break
		}
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
	return nil
}

type OutputVisualizer struct {
	Visualizer
}

type Output struct{}

func (o OutputVisualizer) Output(ctx context.Context, result api.Result) error {
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

		select {
		case <-ctx.Done():
			return nil
		default:
			break
		}

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
		u, err := parseURL(word.Picture)
		if err != nil {
			slog.Error("error processing picture url", "error", err)
			continue
		}
		if u == nil {
			continue
		}
		if err = o.Show(u); err != nil {
			slog.Error("cannot visualize picture", "error", err)
			continue
		}
	}
	return nil
}

// PrintAddedTranslation prints transcription during adding operation
func PrintAddedTranslation(result api.Result) {
	var strTitle string
	if result.InDictionary() {
		strTitle = "Updated existing"
	} else {
		strTitle = "Added new"
	}
	if err := messages.Message(messages.RED, "%s word: ", strTitle); err != nil {
		slog.Error("cannot show message", "error", err)
	}

	if err := messages.Message(messages.GREEN, "['%s'] ['%s']\n", result.Word, strings.Join(result.AddWords, ", ")); err != nil {
		slog.Error("cannot show message", "error", err)
	}
}
