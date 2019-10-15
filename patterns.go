package main

import (
	"context"
	"sync"
)

func orDone(ctx context.Context, input <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
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

func fanOut(ctx context.Context, inputs ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	wg.Add(len(inputs))
	out := make(chan interface{})

	multiplex := func(c <-chan interface{}) {
		defer wg.Done()
		for val := range c {
			select {
			case <-ctx.Done():
				return
			case out <- val:
			}
		}
	}

	for _, c := range inputs {
		go multiplex(c)
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func tee(ctx context.Context, input <-chan interface{}) (_, _ <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})
	go func() {
		defer close(out1)
		defer close(out2)
		for val := range orDone(ctx, input) {
			var out1, out2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case <-ctx.Done():
				case out1 <- val:
					out1 = nil
				case out2 <- val:
					out2 = nil
				}
			}
		}
	}()
	return out1, out2
}

func repeat(ctx context.Context, input ...interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			for _, val := range input {
				select {
				case <-ctx.Done():
					return
				case out <- val:
				}
			}
		}
	}()
	return out
}

func toChannel(ctx context.Context, input ...interface{}) <-chan interface{} {
	out := make(chan interface{})
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

func take(ctx context.Context, input <-chan interface{}, num int) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for i := 0; i < num; i++ {
			select {
			case <-ctx.Done():
				return
			case out <- <-input:
			}
		}
	}()
	return out
}

func bridge(ctx context.Context, inputs <-chan <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			var stream <-chan interface{}
			select {
			case maybe, ok := <-inputs:
				if !ok {
					return
				}
				stream = maybe
			case <-ctx.Done():
				return
			}
			for val := range orDone(ctx, stream) {
				select {
				case out <- val:
				case <-ctx.Done():
				}
			}
		}
	}()
	return out
}

func orderedChannel(input <-chan interface{}, count int) <-chan interface{} {
	out := make(chan interface{}, count)

	go func() {
		defer close(out)
		slideIndex := 0
		results := newIndexedHeap()

		check := func(obj *Indexed) bool {
			return obj != nil && slideIndex == (*obj).getIndex()
		}
		for result := range input {
			results.Add(result.(Indexed))
			for orderResult := results.PullWithCondition(check); orderResult != nil; orderResult = results.PullWithCondition(check) {
				slideIndex++
				out <- *orderResult
			}
		}
		for orderResult := results.Pull(); orderResult != nil; orderResult = results.Pull() {
			out <- *orderResult
		}
	}()

	return out
}
