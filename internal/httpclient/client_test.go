package httpclient

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewUsesDefaults(t *testing.T) {
	t.Parallel()

	client := New(Config{})

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.Equal(t, DefaultMaxIdleConns, transport.MaxIdleConns)
	require.Equal(t, DefaultMaxIdleConnsHost, transport.MaxIdleConnsPerHost)
}

func TestNewUsesProvidedValues(t *testing.T) {
	t.Parallel()

	client := New(Config{MaxIdleConns: 42, MaxIdleConnsPerHost: 7})

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.Equal(t, 42, transport.MaxIdleConns)
	require.Equal(t, 7, transport.MaxIdleConnsPerHost)
}

func TestWithTimeoutSetsCustomContextCause(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	reqWithTimeout, cancel := WithTimeout(context.Background(), req, 10*time.Millisecond)
	defer cancel()

	select {
	case <-reqWithTimeout.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("request context was not canceled by timeout")
	}

	require.True(t, errors.Is(context.Cause(reqWithTimeout.Context()), errRequestTimeout))
}
