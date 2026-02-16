package browser

import (
	"context"
	"net/url"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	visualizer := New()
	require.NotNil(t, visualizer)
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

func TestOpenCommand(t *testing.T) {
	// This test verifies the open function doesn't panic
	// We use a context with timeout to prevent hanging if the browser opens

	testURL, err := url.Parse("https://example.com")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// The open function may fail in CI environments without a display
	// We just verify it doesn't panic and handles the context properly
	_ = open(ctx, testURL)
}

func TestOpenWithValidURL(t *testing.T) {
	// Skip this test on CI or headless environments
	if runtime.GOOS == "linux" && getEnvWithDefault("CI", "") != "" {
		t.Skip("Skipping browser test in CI environment")
	}

	testURL, err := url.Parse("https://example.com/image.png")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// The function may return an error if no display is available,
	// but it shouldn't panic
	_ = open(ctx, testURL)
}

func TestVisualizerShow(t *testing.T) {
	// Skip in CI
	if getEnvWithDefault("CI", "") != "" {
		t.Skip("Skipping browser test in CI environment")
	}

	testURL, err := url.Parse("https://example.com/image.png")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	visualizer := New()
	_ = visualizer.Show(ctx, testURL)
}

// Helper to get environment variable with default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
