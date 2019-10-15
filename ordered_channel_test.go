package main

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestIndexer struct {
	idx int
}

var count = 10

func (ti TestIndexer) getIndex() int {
	return ti.idx
}

func TestOrderedChannel(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	defer done()
	numbers := make([]int, 0, count)
	for i := 0; i < count; i++ {
		numbers = append(numbers, i)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	messages := make([]interface{}, 0, count)
	for i := 0; i < count; i++ {
		messages = append(messages, TestIndexer{idx: numbers[i]})
	}

	channel := orderedChannel(toChannel(ctx, messages...), count)

	min := 0

	for message := range channel {
		idx := message.(TestIndexer).getIndex()
		require.Truef(t, min <= idx, "Value %d is not less or equal than %d", min, idx)
		min = idx
	}

}
