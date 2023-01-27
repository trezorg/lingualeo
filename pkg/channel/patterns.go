package channel

import (
	"context"
)

func ToChannel[V any](ctx context.Context, input ...V) <-chan V {
	out := make(chan V)
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

func OrDone[V any](ctx context.Context, input <-chan V) <-chan V {
	out := make(chan V)
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
