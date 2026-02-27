package translator

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trezorg/lingualeo/internal/api"
)

type translateConcurrencyClient struct {
	started chan struct{}
	release chan struct{}
	current atomic.Int32
	max     atomic.Int32
}

func (c *translateConcurrencyClient) TranslateWord(_ context.Context, word string) api.OperationResult {
	current := c.current.Add(1)
	for {
		maxSeen := c.max.Load()
		if current <= maxSeen {
			break
		}
		if c.max.CompareAndSwap(maxSeen, current) {
			break
		}
	}
	c.started <- struct{}{}
	<-c.release
	c.current.Add(-1)

	return api.OperationResult{Result: api.Result{Word: word}}
}

func (*translateConcurrencyClient) AddWord(_ context.Context, _, _ string) api.OperationResult {
	return api.OperationResult{}
}

type addConcurrencyClient struct {
	started chan struct{}
	release chan struct{}
	current atomic.Int32
	max     atomic.Int32
}

func (*addConcurrencyClient) TranslateWord(_ context.Context, _ string) api.OperationResult {
	return api.OperationResult{}
}

func (c *addConcurrencyClient) AddWord(_ context.Context, word string, translation string) api.OperationResult {
	current := c.current.Add(1)
	for {
		maxSeen := c.max.Load()
		if current <= maxSeen {
			break
		}
		if c.max.CompareAndSwap(maxSeen, current) {
			break
		}
	}
	c.started <- struct{}{}
	<-c.release
	c.current.Add(-1)

	return api.OperationResult{
		Result: api.Result{
			Word:     word,
			AddWords: []string{translation},
		},
	}
}

func TestTranslateWordsRespectsWorkersLimit(t *testing.T) {
	t.Parallel()

	const (
		workers = 2
		count   = 6
	)

	ctx := t.Context()
	client := &translateConcurrencyClient{
		started: make(chan struct{}, count),
		release: make(chan struct{}),
	}
	words := make(chan string, count)
	for i := range count {
		words <- "word"
		_ = i
	}
	close(words)

	ch := translateWords(ctx, client, words, workers)
	waitStarts(t, client.started, workers)
	assertNoAdditionalStart(t, client.started)

	close(client.release)

	results := collectResults(t, ch)
	require.Len(t, results, count)
	require.LessOrEqual(t, int(client.max.Load()), workers)
}

func TestAddWordsRespectsWorkersLimit(t *testing.T) {
	t.Parallel()

	const (
		workers = 2
		count   = 6
	)

	ctx := t.Context()
	client := &addConcurrencyClient{
		started: make(chan struct{}, count),
		release: make(chan struct{}),
	}
	input := make(chan api.Result, count)
	for range count {
		input <- api.Result{Word: "word", AddWords: []string{"translation"}}
	}
	close(input)

	ch := addWords(ctx, client, input, workers)
	waitStarts(t, client.started, workers)
	assertNoAdditionalStart(t, client.started)

	close(client.release)

	results := collectResults(t, ch)
	require.Len(t, results, count)
	require.LessOrEqual(t, int(client.max.Load()), workers)
}

func TestWorkerCountUsesDefaultForNonPositiveValues(t *testing.T) {
	t.Parallel()

	require.Equal(t, defaultWorkers, workerCount(0))
	require.Equal(t, defaultWorkers, workerCount(-1))
	require.Equal(t, 3, workerCount(3))
}

func TestWorkerCountForItemsCapsByItemCount(t *testing.T) {
	t.Parallel()

	require.Equal(t, 2, workerCountForItems(8, 2))
	require.Equal(t, defaultWorkers, workerCountForItems(0, defaultWorkers+2))
	require.Equal(t, 1, workerCountForItems(5, 1))
}

func waitStarts(t *testing.T, started <-chan struct{}, expected int) {
	t.Helper()

	for range expected {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("worker was not started")
		}
	}
}

func assertNoAdditionalStart(t *testing.T, started <-chan struct{}) {
	t.Helper()

	select {
	case <-started:
		t.Fatal("started more workers than expected")
	case <-time.After(150 * time.Millisecond):
	}
}

func collectResults(t *testing.T, ch <-chan api.OperationResult) []api.OperationResult {
	t.Helper()

	done := make(chan []api.OperationResult, 1)
	go func() {
		results := make([]api.OperationResult, 0)
		for res := range ch {
			results = append(results, res)
		}
		done <- results
	}()

	select {
	case results := <-done:
		return results
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for results")
		return nil
	}
}
