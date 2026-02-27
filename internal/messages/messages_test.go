package messages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageSupportsKnownAndFallbackColors(t *testing.T) {
	t.Parallel()

	colors := []Color{RED, GREEN, YELLOW, WHITE, Color(99)}
	for _, c := range colors {
		err := Message(c, "%s", "ok")
		require.NoError(t, err)
	}
}
