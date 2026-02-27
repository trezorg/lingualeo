package translator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendWithContext(t *testing.T) {
	t.Parallel()

	t.Run("active context sends value", func(t *testing.T) {
		t.Parallel()

		out := make(chan int, 1)
		ok := sendWithContext(context.Background(), out, 7)
		require.True(t, ok)
		require.Equal(t, 7, <-out)
	})

	t.Run("canceled context does not send", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(context.Canceled)

		out := make(chan int, 1)
		ok := sendWithContext(ctx, out, 7)
		require.False(t, ok)
		require.Empty(t, out)
	})
}

func TestCheckMediaPlayer(t *testing.T) {
	t.Parallel()

	t.Run("sound disabled", func(t *testing.T) {
		t.Parallel()

		l := Lingualeo{Sound: false}
		require.NoError(t, l.checkMediaPlayer())
		require.False(t, l.Sound)
	})

	t.Run("empty player disables sound", func(t *testing.T) {
		t.Parallel()

		l := Lingualeo{Sound: true}
		err := l.checkMediaPlayer()
		require.Error(t, err)
		require.Contains(t, err.Error(), "player parameter not set")
		require.False(t, l.Sound)
	})

	t.Run("unavailable player disables sound", func(t *testing.T) {
		t.Parallel()

		l := Lingualeo{Sound: true, Player: "definitely-not-existing-command-123456"}
		err := l.checkMediaPlayer()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not available")
		require.False(t, l.Sound)
	})
}

func TestOutputerReturnsUnknownVisualizeTypeError(t *testing.T) {
	t.Parallel()

	_, err := outputer(true, VisualiseType("broken"))
	require.Error(t, err)
	require.True(t, errors.Is(err, errUnknownVisualiseType))
}
