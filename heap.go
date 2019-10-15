package main

import (
	"container/heap"
	"sync"
)

type indexedHeap struct {
	messages []indexedItem
	lock     sync.RWMutex
}

func (h *indexedHeap) Len() int           { return len(h.messages) }
func (h *indexedHeap) Less(i, j int) bool { return h.messages[i].getIndex() < h.messages[j].getIndex() }
func (h *indexedHeap) Swap(i, j int)      { h.messages[i], h.messages[j] = h.messages[j], h.messages[i] }

// Push and Pop use pointer receivers because they modify the slice's length,
// not just its contents.
func (h *indexedHeap) Push(x interface{}) {
	(*h).messages = append((*h).messages, x.(indexedItem))
}

func (h *indexedHeap) Pop() interface{} {
	old := (*h).messages
	n := len(old)
	x := old[n-1]
	(*h).messages = old[0 : n-1]
	return x
}

func (h *indexedHeap) Add(message indexedItem) {
	h.lock.Lock()
	defer h.lock.Unlock()
	heap.Push(h, message)
}

func (h *indexedHeap) AddMany(messages ...indexedItem) {
	h.lock.Lock()
	defer h.lock.Unlock()
	for _, message := range messages {
		heap.Push(h, message)
	}
}

func (h *indexedHeap) Pull() *indexedItem {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.Len() == 0 {
		return nil
	}
	message := heap.Pop(h).(indexedItem)
	return &message
}

func (h *indexedHeap) PullWithCondition(check func(*indexedItem) bool) *indexedItem {
	h.lock.Lock()
	defer h.lock.Unlock()
	if check(h.pick()) {
		message := heap.Pop(h).(indexedItem)
		return &message
	}
	return nil
}

func (h *indexedHeap) Pick() *indexedItem {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.pick()
}

func (h *indexedHeap) pick() *indexedItem {
	if h.Len() > 0 {
		message := (*h).messages[0]
		return &message
	}
	return nil
}

func newIndexedHeap() *indexedHeap {
	h := &indexedHeap{}
	heap.Init(h)
	return h
}
