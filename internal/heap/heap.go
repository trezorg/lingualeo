package heap

import (
	"container/heap"
	"sync"
)

type IndexedItem interface {
	GetIndex() int
}

type IndexedHeap struct {
	messages []IndexedItem
	lock     sync.RWMutex
}

func (h *IndexedHeap) Len() int           { return len(h.messages) }
func (h *IndexedHeap) Less(i, j int) bool { return h.messages[i].GetIndex() < h.messages[j].GetIndex() }
func (h *IndexedHeap) Swap(i, j int)      { h.messages[i], h.messages[j] = h.messages[j], h.messages[i] }

// Push and Pop use pointer receivers because they modify the slice's length,
// not just its contents.
func (h *IndexedHeap) Push(x any) {
	h.messages = append(h.messages, x.(IndexedItem))
}

func (h *IndexedHeap) Pop() any {
	old := h.messages
	n := len(old)
	x := old[n-1]
	h.messages = old[0 : n-1]
	return x
}

func (h *IndexedHeap) Add(message IndexedItem) {
	h.lock.Lock()
	defer h.lock.Unlock()
	heap.Push(h, message)
}

func (h *IndexedHeap) AddMany(messages ...IndexedItem) {
	h.lock.Lock()
	defer h.lock.Unlock()
	for _, message := range messages {
		heap.Push(h, message)
	}
}

func (h *IndexedHeap) Pull() *IndexedItem {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.Len() == 0 {
		return nil
	}
	raw := heap.Pop(h)
	if message, ok := raw.(IndexedItem); ok {
		return &message
	}
	return nil
}

func (h *IndexedHeap) PullWithCondition(check func(*IndexedItem) bool) *IndexedItem {
	h.lock.Lock()
	defer h.lock.Unlock()
	if check(h.peek()) {
		raw := heap.Pop(h)
		if message, ok := raw.(IndexedItem); ok {
			return &message
		}
	}
	return nil
}

func (h *IndexedHeap) Pick() *IndexedItem {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.peek()
}

func (h *IndexedHeap) peek() *IndexedItem {
	if h.Len() > 0 {
		message := h.messages[0]
		return &message
	}
	return nil
}

func NewIndexedHeap() *IndexedHeap {
	h := &IndexedHeap{}
	heap.Init(h)
	return h
}
