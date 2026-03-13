package translator

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trezorg/lingualeo/internal/api"
)

type reverseClient struct {
	mu         sync.Mutex
	translated []string
	firstPass  api.OperationResult
	secondPass api.OperationResult
}

func (c *reverseClient) TranslateWord(_ context.Context, word string) api.OperationResult {
	c.mu.Lock()
	c.translated = append(c.translated, word)
	c.mu.Unlock()
	if word == "привет" {
		return c.firstPass
	}

	return c.secondPass
}

func (*reverseClient) AddWord(_ context.Context, _, _ string) api.OperationResult {
	return api.OperationResult{}
}

func (*reverseClient) Auth(_ context.Context) error {
	return nil
}

type outputCollector struct {
	mu    sync.Mutex
	words []string
}

func (o *outputCollector) Output(_ context.Context, result api.Result) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.words = append(o.words, result.Word)
	return nil
}

func TestTranslateWithReverseRussianKeepsOriginalWords(t *testing.T) {
	t.Parallel()

	client := &reverseClient{
		firstPass: api.OperationResult{
			Result: api.Result{
				Word: "привет",
				Translate: []api.Word{
					{Value: "hello"},
					{Value: "hello"},
				},
			},
		},
		secondPass: api.OperationResult{
			Result: api.Result{
				Word:      "hello",
				Translate: []api.Word{{Value: "привет"}},
			},
		},
	}
	output := &outputCollector{}
	app := Lingualeo{
		Client:   client,
		Outputer: output,
		Config: Config{
			ReverseTranslate: true,
		},
		Words: []string{"привет"},
	}

	app.TranslateWithReverseRussian(t.Context())

	require.Equal(t, []string{"привет"}, app.Words)
	require.Equal(t, []string{"привет", "hello"}, client.translated)
	require.Equal(t, []string{"привет", "hello"}, output.words)
}
