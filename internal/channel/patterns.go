package channel

import (
	"context"
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
