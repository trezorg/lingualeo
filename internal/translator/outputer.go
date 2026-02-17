package translator

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/trezorg/lingualeo/internal/api"
	"github.com/trezorg/lingualeo/internal/messages"
	"github.com/trezorg/lingualeo/internal/validator"
)

type Outputer interface {
	Output(ctx context.Context, r api.Result) error
}

var errParsePictureURL = errors.New("cannot parse picture url")
var errShowMessage = errors.New("cannot show message")

func messagef(c messages.Color, message string, params ...any) error {
	if err := messages.Message(c, message, params...); err != nil {
		return fmt.Errorf("%w: %w", errShowMessage, err)
	}
	return nil
}

func parseURL(s string) (*url.URL, error) {
	if s == "" {
		return nil, nil
	}
	if err := validator.ValidateURL(s); err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errParsePictureURL, s, err)
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errParsePictureURL, s)
	}
	return u, nil
}

func printTranslation(ctx context.Context, result api.Result) error {
	var strTitle string
	if result.InDictionary() {
		strTitle = "existing"
	} else {
		strTitle = "new"
	}
	if err := messagef(messages.RED, "Found %s word:\n", strTitle); err != nil {
		return err
	}
	if err := messagef(messages.GREEN, "['%s'] (%s)\n", result.Word, result.Transcription); err != nil {
		return err
	}
	for _, word := range result.Translate {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := messagef(messages.YELLOW, "%s", word.Value); err != nil {
			return err
		}
		if len(word.Context) > 0 {
			if err := messagef(messages.WHITE, " (%s)", word.Context); err != nil {
				return err
			}
		}
		if err := messagef(messages.YELLOW, "\n"); err != nil {
			return err
		}
	}

	return nil
}

func (Output) Output(ctx context.Context, result api.Result) error {
	return printTranslation(ctx, result)
}

type OutputVisualizer struct {
	Visualizer
}

type Output struct{}

func (o OutputVisualizer) Output(ctx context.Context, result api.Result) error {
	if err := printTranslation(ctx, result); err != nil {
		return err
	}

	var outErr error
	for _, word := range result.Translate {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		u, err := parseURL(word.Picture)
		if err != nil {
			outErr = errors.Join(outErr, err)
			continue
		}
		if u == nil {
			continue
		}
		if err = o.Show(ctx, u); err != nil {
			outErr = errors.Join(outErr, err)
			continue
		}
	}
	return outErr
}

// PrintAddedTranslation prints transcription during adding operation
func PrintAddedTranslation(result api.Result) error {
	var strTitle string
	if result.InDictionary() {
		strTitle = "Updated existing"
	} else {
		strTitle = "Added new"
	}
	if err := messagef(messages.RED, "%s word: ", strTitle); err != nil {
		return err
	}

	return messagef(messages.GREEN, "['%s'] ['%s']\n", result.Word, strings.Join(result.AddWords, ", "))
}
