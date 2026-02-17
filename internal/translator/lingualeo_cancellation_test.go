package translator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trezorg/lingualeo/internal/api"
)

type blockingClient struct {
	translateStarted chan struct{}
	translateRelease chan struct{}
	addStarted       chan struct{}
	addRelease       chan struct{}
}

func (c *blockingClient) TranslateWord(_ context.Context, word string) api.OperationResult {
	close(c.translateStarted)
	<-c.translateRelease
	return api.OperationResult{Result: api.Result{Word: word}}
}

func (c *blockingClient) AddWord(_ context.Context, word string, translate string) api.OperationResult {
	close(c.addStarted)
	<-c.addRelease
	return api.OperationResult{
		Result: api.Result{
			Word:     word,
			AddWords: []string{translate},
		},
	}
}

func TestTranslateWordsStopsOnCancelWithoutConsumer(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	client := &blockingClient{
		translateStarted: make(chan struct{}),
		translateRelease: make(chan struct{}),
		addStarted:       make(chan struct{}),
		addRelease:       make(chan struct{}),
	}
	words := make(chan string, 1)
	words <- "word"
	close(words)

	ch := translateWords(ctx, client, words)

	select {
	case <-client.translateStarted:
	case <-time.After(time.Second):
		t.Fatal("translate worker was not started")
	}

	cancel()
	close(client.translateRelease)

	select {
	case _, ok := <-ch:
		require.False(t, ok, "result channel should be closed after cancellation")
	case <-time.After(time.Second):
		t.Fatal("translateWords output channel was not closed after cancellation")
	}
}

func TestAddWordsStopsOnCancelWithoutConsumer(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	client := &blockingClient{
		translateStarted: make(chan struct{}),
		translateRelease: make(chan struct{}),
		addStarted:       make(chan struct{}),
		addRelease:       make(chan struct{}),
	}
	words := make(chan api.Result, 1)
	words <- api.Result{
		Word:     "word",
		AddWords: []string{"translation"},
	}
	close(words)

	ch := addWords(ctx, client, words)

	select {
	case <-client.addStarted:
	case <-time.After(time.Second):
		t.Fatal("add worker was not started")
	}

	cancel()
	close(client.addRelease)

	select {
	case _, ok := <-ch:
		require.False(t, ok, "result channel should be closed after cancellation")
	case <-time.After(time.Second):
		t.Fatal("addWords output channel was not closed after cancellation")
	}
}
