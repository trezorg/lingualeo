package player

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		player     string
		wantExec   string
		wantParams []string
	}{
		{
			name:       "simple player",
			player:     "mpv",
			wantExec:   "mpv",
			wantParams: []string{},
		},
		{
			name:       "player with params",
			player:     "mpv --really-quiet",
			wantExec:   "mpv",
			wantParams: []string{"--really-quiet"},
		},
		{
			name:       "player with multiple params",
			player:     "mpv --really-quiet --no-video",
			wantExec:   "mpv",
			wantParams: []string{"--really-quiet", "--no-video"},
		},
		{
			name:       "empty player",
			player:     "",
			wantExec:   "",
			wantParams: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.player)
			assert.Equal(t, tt.wantExec, p.player)
			assert.Equal(t, tt.wantParams, p.params)
		})
	}
}

func TestPlayerPlay(t *testing.T) {
	tests := []struct {
		name        string
		player      string
		url         string
		wantErr     bool
		skipCommand bool
	}{
		{
			name:   "play with echo command",
			player: "echo",
			url:    "https://example.com/audio.mp3",
		},
		{
			name:   "play with true command",
			player: "true",
			url:    "https://example.com/audio.mp3",
		},
		{
			name:    "non-existent command",
			player:  "nonexistentcommand12345",
			url:     "https://example.com/audio.mp3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.player)
			ctx := context.Background()

			err := p.Play(ctx, tt.url)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPlayerPlayWithCancellation(t *testing.T) {
	// Use sleep command that runs for a while, then cancel
	p := New("sleep 10")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := p.Play(ctx, "https://example.com/audio.mp3")
	elapsed := time.Since(start)

	// Should return quickly due to context cancellation (within shutdown timeout + margin)
	require.True(t, elapsed < 3*time.Second, "Play should return quickly after context cancellation")
	// The error will be either context deadline exceeded or process killed
	require.Error(t, err)
}

func TestPlayerPlayWithEmptyPlayer(t *testing.T) {
	p := New("")
	ctx := context.Background()

	err := p.Play(ctx, "https://example.com/audio.mp3")
	require.Error(t, err)
}

func TestPlayerPlayParamsAreNotModified(t *testing.T) {
	p := New("echo test")
	originalParams := make([]string, len(p.params))
	copy(originalParams, p.params)

	ctx := context.Background()
	_ = p.Play(ctx, "https://example.com/audio.mp3")

	// Verify params are not modified
	assert.Equal(t, originalParams, p.params)
}

func TestPlayerStruct(t *testing.T) {
	p := Player{
		player: "test-player",
		params: []string{"-a", "-b"},
	}

	assert.Equal(t, "test-player", p.player)
	assert.Equal(t, []string{"-a", "-b"}, p.params)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, " ", separator)
	assert.Equal(t, 2*time.Second, defaultShutdownTimeout)
}

func TestWithShutdownTimeout(t *testing.T) {
	p := New("echo", WithShutdownTimeout(5*time.Second))
	assert.Equal(t, 5*time.Second, p.shutdownTimeout)
}

func TestNewWithDefaults(t *testing.T) {
	p := New("echo")
	assert.Equal(t, "echo", p.player)
	assert.Equal(t, defaultShutdownTimeout, p.shutdownTimeout)
}
