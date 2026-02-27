package slice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUniquePreservesOrder(t *testing.T) {
	t.Parallel()

	in := []int{3, 1, 3, 2, 1, 2}
	require.Equal(t, []int{3, 1, 2}, Unique(in))
}

func TestUniqueFuncPreservesOrder(t *testing.T) {
	t.Parallel()

	type item struct {
		key   string
		value int
	}

	in := []item{{key: "a", value: 1}, {key: "b", value: 2}, {key: "a", value: 3}}
	out := UniqueFunc(in, func(i item) string { return i.key })

	require.Equal(t, []item{{key: "a", value: 1}, {key: "b", value: 2}}, out)
}
