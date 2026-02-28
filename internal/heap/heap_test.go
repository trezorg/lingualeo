package heap

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testItem implements IndexedItem for testing
type testItem struct {
	value int
	index int
}

func (t testItem) GetIndex() int {
	return t.index
}

func newTestItem(value, index int) testItem {
	return testItem{value: value, index: index}
}

func TestNewIndexedHeap(t *testing.T) {
	h := NewIndexedHeap()
	assert.NotNil(t, h)
	assert.Equal(t, 0, h.Len())
}

func TestIndexedHeap_Add(t *testing.T) {
	t.Run("single item", func(t *testing.T) {
		h := NewIndexedHeap()
		h.Add(newTestItem(10, 1))
		assert.Equal(t, 1, h.Len())
	})

	t.Run("multiple items", func(t *testing.T) {
		h := NewIndexedHeap()
		h.Add(newTestItem(10, 3))
		h.Add(newTestItem(20, 1))
		h.Add(newTestItem(30, 2))
		assert.Equal(t, 3, h.Len())
	})
}

func TestIndexedHeap_AddMany(t *testing.T) {
	h := NewIndexedHeap()
	h.AddMany(newTestItem(10, 3), newTestItem(20, 1), newTestItem(30, 2))
	assert.Equal(t, 3, h.Len())
}

func TestIndexedHeap_Pull(t *testing.T) {
	t.Run("empty heap", func(t *testing.T) {
		h := NewIndexedHeap()
		item := h.Pull()
		assert.Nil(t, item)
	})

	t.Run("single item", func(t *testing.T) {
		h := NewIndexedHeap()
		h.Add(newTestItem(42, 1))
		item := h.Pull()
		assert.NotNil(t, item)
		assert.Equal(t, 1, (*item).GetIndex())
		assert.Equal(t, 0, h.Len())
	})

	t.Run("ordered by index", func(t *testing.T) {
		h := NewIndexedHeap()
		h.Add(newTestItem(10, 5))
		h.Add(newTestItem(20, 2))
		h.Add(newTestItem(30, 8))
		h.Add(newTestItem(40, 1))

		item := h.Pull()
		assert.NotNil(t, item)
		assert.Equal(t, 1, (*item).GetIndex())

		item = h.Pull()
		assert.NotNil(t, item)
		assert.Equal(t, 2, (*item).GetIndex())

		item = h.Pull()
		assert.NotNil(t, item)
		assert.Equal(t, 5, (*item).GetIndex())

		item = h.Pull()
		assert.NotNil(t, item)
		assert.Equal(t, 8, (*item).GetIndex())

		item = h.Pull()
		assert.Nil(t, item)
	})
}

func TestIndexedHeap_PullWithCondition(t *testing.T) {
	t.Run("empty heap", func(t *testing.T) {
		h := NewIndexedHeap()
		item := h.PullWithCondition(func(item *IndexedItem) bool {
			return true
		})
		assert.Nil(t, item)
	})

	t.Run("condition not met", func(t *testing.T) {
		h := NewIndexedHeap()
		h.Add(newTestItem(10, 5))
		item := h.PullWithCondition(func(item *IndexedItem) bool {
			return item != nil && (*item).GetIndex() == 1
		})
		assert.Nil(t, item)
		assert.Equal(t, 1, h.Len()) // Item not removed
	})

	t.Run("condition met", func(t *testing.T) {
		h := NewIndexedHeap()
		h.Add(newTestItem(10, 1))
		h.Add(newTestItem(20, 2))
		h.Add(newTestItem(30, 3))

		item := h.PullWithCondition(func(item *IndexedItem) bool {
			return item != nil && (*item).GetIndex() == 1
		})
		assert.NotNil(t, item)
		assert.Equal(t, 1, (*item).GetIndex())
		assert.Equal(t, 2, h.Len())
	})

	t.Run("sequential pulls", func(t *testing.T) {
		h := NewIndexedHeap()
		h.AddMany(newTestItem(10, 1), newTestItem(20, 2), newTestItem(30, 3))

		slideIndex := 0
		check := func(item *IndexedItem) bool {
			return item != nil && slideIndex == (*item).GetIndex()
		}

		// Pull index 0 - should fail
		item := h.PullWithCondition(check)
		assert.Nil(t, item)

		// Pull index 1 - should succeed
		slideIndex = 1
		item = h.PullWithCondition(check)
		assert.NotNil(t, item)
		assert.Equal(t, 1, (*item).GetIndex())

		// Pull index 2 - should succeed
		slideIndex = 2
		item = h.PullWithCondition(check)
		assert.NotNil(t, item)
		assert.Equal(t, 2, (*item).GetIndex())

		// Pull index 3 - should succeed
		slideIndex = 3
		item = h.PullWithCondition(check)
		assert.NotNil(t, item)
		assert.Equal(t, 3, (*item).GetIndex())

		assert.Equal(t, 0, h.Len())
	})
}

func TestIndexedHeap_Pick(t *testing.T) {
	t.Run("empty heap", func(t *testing.T) {
		h := NewIndexedHeap()
		item := h.Pick()
		assert.Nil(t, item)
	})

	t.Run("non-empty heap", func(t *testing.T) {
		h := NewIndexedHeap()
		h.Add(newTestItem(10, 5))
		h.Add(newTestItem(20, 1))
		h.Add(newTestItem(30, 3))

		item := h.Pick()
		assert.NotNil(t, item)
		assert.Equal(t, 1, (*item).GetIndex()) // Should return smallest index
		assert.Equal(t, 3, h.Len())            // Item not removed
	})
}

func TestIndexedHeap_Concurrency(t *testing.T) {
	t.Run("concurrent adds", func(t *testing.T) {
		h := NewIndexedHeap()
		var wg sync.WaitGroup

		for i := range 100 {
			wg.Go(func() {
				h.Add(newTestItem(i, i))
			})
		}

		wg.Wait()
		assert.Equal(t, 100, h.Len())
	})

	t.Run("concurrent add and pull", func(t *testing.T) {
		h := NewIndexedHeap()
		var addWg sync.WaitGroup

		// Add items first
		for i := range 50 {
			addWg.Go(func() {
				h.Add(newTestItem(i, i))
			})
		}

		// Wait for all adds to complete
		addWg.Wait()

		// Now pull items concurrently
		var pullWg sync.WaitGroup
		pulled := make([]int, 0, 50)
		var pullMu sync.Mutex
		for range 50 {
			pullWg.Go(func() {
				item := h.Pull()
				if item != nil {
					pullMu.Lock()
					pulled = append(pulled, (*item).GetIndex())
					pullMu.Unlock()
				}
			})
		}

		pullWg.Wait()
		assert.Len(t, pulled, 50)
	})

	t.Run("concurrent add many and pick", func(t *testing.T) {
		h := NewIndexedHeap()
		var wg sync.WaitGroup

		for i := range 10 {
			wg.Go(func() {
				items := make([]IndexedItem, 10)
				for j := range 10 {
					items[j] = newTestItem(i*10+j, i*10+j)
				}
				h.AddMany(items...)
			})
		}

		// Concurrent picks
		for range 20 {
			wg.Go(func() {
				_ = h.Pick()
			})
		}

		wg.Wait()
		assert.Equal(t, 100, h.Len())
	})
}

func TestIndexedHeap_Ordering(t *testing.T) {
	t.Run("maintains min-heap property", func(t *testing.T) {
		h := NewIndexedHeap()

		// Add items in random order
		indices := []int{7, 3, 9, 1, 5, 2, 8, 4, 6, 0}
		for _, idx := range indices {
			h.Add(newTestItem(idx*10, idx))
		}

		// Pull should return items in ascending index order
		for i := range 10 {
			item := h.Pull()
			assert.NotNil(t, item)
			assert.Equal(t, i, (*item).GetIndex(), "expected index %d", i)
		}
	})
}
