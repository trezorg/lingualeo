package channel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToChannel(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		ctx := t.Context()
		ch := ToChannel[int](ctx)
		assert.Empty(t, drainChannel(ch))
	})

	t.Run("single item", func(t *testing.T) {
		ctx := t.Context()
		ch := ToChannel(ctx, 1)
		assert.Equal(t, []int{1}, drainChannel(ch))
	})

	t.Run("multiple items", func(t *testing.T) {
		ctx := t.Context()
		ch := ToChannel(ctx, 1, 2, 3, 4, 5)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, drainChannel(ch))
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(t.Context())
		cancel(context.Canceled) // Cancel immediately
		ch := ToChannel(ctx, 1, 2, 3)
		// Give goroutine time to process
		time.Sleep(10 * time.Millisecond)
		// Channel should be closed without sending all items
		_, ok := <-ch
		assert.False(t, ok, "channel should be closed after context cancellation")
	})

	t.Run("string type", func(t *testing.T) {
		ctx := t.Context()
		ch := ToChannel(ctx, "hello", "world")
		assert.Equal(t, []string{"hello", "world"}, drainChannel(ch))
	})
}

func TestOrDone(t *testing.T) {
	t.Run("empty channel", func(t *testing.T) {
		ctx := t.Context()
		input := make(chan int)
		close(input)
		ch := OrDone(ctx, input)
		assert.Empty(t, drainChannel(ch))
	})

	t.Run("items passed through", func(t *testing.T) {
		ctx := t.Context()
		input := make(chan int, 3)
		input <- 1
		input <- 2
		input <- 3
		close(input)
		ch := OrDone(ctx, input)
		assert.Equal(t, []int{1, 2, 3}, drainChannel(ch))
	})

	t.Run("context cancellation mid-stream", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(t.Context())
		input := make(chan int)

		// Start a goroutine that will send items slowly
		go func() {
			input <- 1
			time.Sleep(50 * time.Millisecond)
			input <- 2 // This may or may not be sent depending on timing
			close(input)
		}()

		ch := OrDone(ctx, input)

		// Read first item
		val := <-ch
		assert.Equal(t, 1, val)

		// Cancel context
		cancel(context.Canceled)
		time.Sleep(20 * time.Millisecond)

		// Channel should close after context cancellation
		_, ok := <-ch
		assert.False(t, ok, "channel should be closed after context cancellation")
	})

	t.Run("large number of items", func(t *testing.T) {
		ctx := t.Context()
		input := make(chan int, 100)
		for i := range 100 {
			input <- i
		}
		close(input)
		ch := OrDone(ctx, input)
		results := drainChannel(ch)
		assert.Len(t, results, 100)
	})

	t.Run("already canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(t.Context())
		cancel(context.Canceled)
		input := make(chan int, 1)
		input <- 1
		close(input)
		ch := OrDone(ctx, input)
		// Channel should close immediately due to canceled context
		time.Sleep(10 * time.Millisecond)
		_, ok := <-ch
		assert.False(t, ok, "channel should be closed with canceled context")
	})
}

func drainChannel[T any](ch <-chan T) []T {
	var result []T
	for val := range ch {
		result = append(result, val)
	}
	return result
}
