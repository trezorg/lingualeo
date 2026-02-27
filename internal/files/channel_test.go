package files

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedChannelOrdersByIndex(t *testing.T) {
	t.Parallel()

	in := make(chan File, 3)
	in <- File{Filename: "third", Index: 2}
	in <- File{Filename: "first", Index: 0}
	in <- File{Filename: "second", Index: 1}
	close(in)

	out := OrderedChannel(in, 3)

	got := make([]string, 0, 3)
	for f := range out {
		got = append(got, f.Filename)
	}

	require.Equal(t, []string{"first", "second", "third"}, got)
}
