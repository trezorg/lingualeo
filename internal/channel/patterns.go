package channel

import (
	"context"
	"sync"
	"time"
)

func ToChannel[T any](ctx context.Context, input ...T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, val := range input {
			select {
			case <-ctx.Done():
				return
			case out <- val:
			}
		}
	}()
	return out
}

func OrDone[T any](ctx context.Context, input <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-input:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
				}
			}
		}
	}()
	return out
}

// Merge combines multiple input channels into a single output channel.
// The output channel is closed when all input channels are closed.
func Merge[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	for _, ch := range inputs {
		wg.Go(func() {
			for v := range OrDone(ctx, ch) {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Batch collects items into slices of up to maxSize.
// Flushes immediately when maxSize is reached or on channel close.
func Batch[T any](ctx context.Context, input <-chan T, maxSize int) <-chan []T {
	out := make(chan []T)
	go func() {
		defer close(out)
		batch := make([]T, 0, maxSize)
		for v := range OrDone(ctx, input) {
			batch = append(batch, v)
			if len(batch) >= maxSize {
				select {
				case out <- batch:
				case <-ctx.Done():
					return
				}
				batch = make([]T, 0, maxSize)
			}
		}
		// Flush remaining items
		if len(batch) > 0 {
			select {
			case out <- batch:
			case <-ctx.Done():
			}
		}
	}()
	return out
}

// BatchWithTimeout collects items into slices, flushing after maxItems
// or after timeout duration since first item in batch.
func BatchWithTimeout[T any](ctx context.Context, input <-chan T, maxItems int, timeout time.Duration) <-chan []T {
	out := make(chan []T)
	go func() {
		defer close(out)
		var (
			batch     []T
			timer     *time.Timer
			timerChan <-chan time.Time
		)
		flush := func() {
			if len(batch) > 0 {
				select {
				case out <- batch:
				case <-ctx.Done():
				}
			}
			batch = nil
			if timer != nil {
				timer.Stop()
				timer = nil
				timerChan = nil
			}
		}
		for {
			select {
			case <-ctx.Done():
				flush()
				return
			case <-timerChan:
				flush()
			case v, ok := <-input:
				if !ok {
					flush()
					return
				}
				if batch == nil {
					batch = make([]T, 0, maxItems)
					timer = time.NewTimer(timeout)
					timerChan = timer.C
				}
				batch = append(batch, v)
				if len(batch) >= maxItems {
					flush()
				}
			}
		}
	}()
	return out
}
