package testing

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trezorg/lingualeo/internal/files"
)

var count = 1000

func TestOrderedChannel(t *testing.T) {

	numbers := make([]int, 0, count)
	for i := 0; i < count; i++ {
		numbers = append(numbers, i)
	}
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	messages := make(chan files.File, count)
	for i := 0; i < count; i++ {
		messages <- files.File{
			Error:    nil,
			Filename: "",
			Index:    numbers[i],
		}
	}
	close(messages)
	require.Equal(t, count, len(messages), "Channel should have size: %d. But has: %d", count, len(messages))

	chn := files.OrderedChannel(messages, count)
	min := 0

	for message := range chn {
		idx := message.GetIndex()
		require.Truef(t, min <= idx, "Value %d is not less or equal than %d", min, idx)
		min = idx
	}
}
