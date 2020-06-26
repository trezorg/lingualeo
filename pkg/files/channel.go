package files

import (
	"context"

	"github.com/trezorg/lingualeo/pkg/heap"
)

func OrderedChannel(input <-chan File, count int) <-chan File {
	out := make(chan File, count)

	go func() {
		defer close(out)
		slideIndex := 0
		results := heap.NewIndexedHeap()

		check := func(obj *heap.IndexedItem) bool {
			return obj != nil && slideIndex == (*obj).GetIndex()
		}
		for result := range input {
			results.Add(result)
			for orderResult := results.PullWithCondition(check); orderResult != nil; orderResult = results.PullWithCondition(check) {
				slideIndex++
				out <- (*orderResult).(File)
			}
		}
		for orderResult := results.Pull(); orderResult != nil; orderResult = results.Pull() {
			out <- (*orderResult).(File)
		}
	}()

	return out
}

func OrFilesDone(ctx context.Context, input <-chan File) <-chan File {
	out := make(chan File)
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
