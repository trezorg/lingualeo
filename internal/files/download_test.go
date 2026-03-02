package files

import (
	"context"
	"net/http"
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
