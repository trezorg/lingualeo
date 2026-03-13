package files

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	d := New(http.DefaultClient)
	_, err := d.Download(context.Background(), "not-a-url")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid download URL")
}

func TestDownloadBytesRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	d := New(http.DefaultClient)
	_, err := d.DownloadBytes(context.Background(), "not-a-url")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid download URL")
}

func TestDownloadWritesResponseToTempFile(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("payload"))
	}))
	defer server.Close()

	d := New(server.Client())
	filename, err := d.Download(t.Context(), server.URL)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Remove(filename)
	})

	data, readErr := os.ReadFile(filename)
	require.NoError(t, readErr)
	require.Equal(t, []byte("payload"), data)
}

func TestDownloadBytesReadsResponseBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("payload"))
	}))
	defer server.Close()

	d := New(server.Client())
	data, err := d.DownloadBytes(t.Context(), server.URL)
	require.NoError(t, err)
	require.Equal(t, []byte("payload"), data)
}
