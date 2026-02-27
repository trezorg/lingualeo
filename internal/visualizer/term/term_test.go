package term

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMode(t *testing.T) {
	// Mode() returns a GraphicMode based on terminal capabilities
	// This test just verifies it doesn't panic and returns a valid mode
	mode := Mode()
	assert.NotEmpty(t, mode)

	// The mode should be one of the defined constants
	validModes := map[GraphicMode]bool{
		Sixel:   true,
		Iterm:   true,
		Kitty:   true,
		Unknown: true,
	}
	assert.True(t, validModes[mode], "Mode should return a valid GraphicMode")
}

func TestNew(t *testing.T) {
	visualizer := New()
	require.NotNil(t, visualizer)
}

func TestVisualizerShow(t *testing.T) {
	// Create a test server that returns a small valid image
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A minimal 1x1 PNG image
		pngData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
			0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
			0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
			0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
			0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, // IEND chunk
			0x44, 0xAE, 0x42, 0x60, 0x82,
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngData)
	}))
	defer server.Close()

	testURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	visualizer := New()

	// The function may or may not display the image depending on terminal support,
	// but it shouldn't return an error for a valid image
	err = visualizer.Show(ctx, testURL)
	// In Unknown mode, the function returns nil without doing anything
	// In other modes, it may succeed or fail depending on terminal support
	// We just verify it doesn't panic
	_ = err
}

func TestOpenWithBadURL(t *testing.T) {
	// Test with a URL that will return an error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	testURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	err = open(ctx, testURL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad status")
}

func TestOpenWithInvalidURL(t *testing.T) {
	// Test with an unreachable URL
	testURL, err := url.Parse("http://localhost:1/invalid")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeoutCause(context.Background(), 100, context.DeadlineExceeded)
	defer cancel()

	err = open(ctx, testURL)
	require.Error(t, err)
}

func TestGraphicModeString(t *testing.T) {
	tests := []struct {
		mode    GraphicMode
		wantStr string
	}{
		{Sixel, "sixel"},
		{Iterm, "iterm"},
		{Kitty, "kitty"},
		{Unknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			assert.Equal(t, tt.wantStr, string(tt.mode))
		})
	}
}

func TestVisualizerType(t *testing.T) {
	// Verify that Visualizer is a function type
	var v Visualizer = func(_ context.Context, _ *url.URL) error {
		return nil
	}
	require.NotNil(t, v)

	// Verify Show method
	err := v.Show(context.Background(), &url.URL{})
	require.NoError(t, err)
}
