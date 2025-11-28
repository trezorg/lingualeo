package files

import (
	"github.com/trezorg/lingualeo/pkg/heap"
)

// OrderedChannel gets channel and returns channel ordered by GetIndex
func OrderedChannel(input <-chan File, count int) <-chan File {
	out := make(chan File, count)

	go func() {
		defer close(out)
		slideIndex := 0
		results := heap.NewIndexedHeap()

		check := func(obj *heap.IndexedItem) bool {
			return obj != nil && slideIndex == (*obj).GetIndex()
		}
		sendResult := func(orderResult *heap.IndexedItem) {
			if orderResult == nil {
				return
			}
			if file, ok := (*orderResult).(File); ok {
				out <- file
			}
		}
		for result := range input {
			results.Add(result)
			for orderResult := results.PullWithCondition(check); orderResult != nil; orderResult = results.PullWithCondition(check) {
				slideIndex++
				sendResult(orderResult)
			}
		}
		for orderResult := results.Pull(); orderResult != nil; orderResult = results.Pull() {
			sendResult(orderResult)
		}
	}()

	return out
}
