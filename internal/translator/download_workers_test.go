package translator

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trezorg/lingualeo/internal/files"
)

type downloadConcurrencyDownloader struct {
	started chan struct{}
	release chan struct{}
	current atomic.Int32
	max     atomic.Int32
}

func (d *downloadConcurrencyDownloader) Download(_ context.Context, url string) (string, error) {
	current := d.current.Add(1)
	for {
		maxSeen := d.max.Load()
		if current <= maxSeen {
			break
		}
		if d.max.CompareAndSwap(maxSeen, current) {
			break
		}
	}
	d.started <- struct{}{}
	<-d.release
	d.current.Add(-1)

	return url, nil
}

func (*downloadConcurrencyDownloader) Remove(_ string) error {
	return nil
}

func TestDownloadFilesRespectsWorkersLimit(t *testing.T) {
	t.Parallel()

	const (
		workers = 2
		count   = 6
	)

	ctx := t.Context()
	downloader := &downloadConcurrencyDownloader{
		started: make(chan struct{}, count),
		release: make(chan struct{}),
	}
	urls := make(chan string, count)
	for range count {
		urls <- "http://example.com/file"
	}
	close(urls)

	ch := downloadFiles(ctx, urls, downloader, workers)
	waitStarts(t, downloader.started, workers)
	assertNoAdditionalStart(t, downloader.started)

	close(downloader.release)

	results := collectFiles(t, ch)
	require.Len(t, results, count)
	require.LessOrEqual(t, int(downloader.max.Load()), workers)
}

func collectFiles(t *testing.T, ch <-chan files.File) []files.File {
	t.Helper()

	done := make(chan []files.File, 1)
	go func() {
		results := make([]files.File, 0)
		for res := range ch {
			results = append(results, res)
		}
		done <- results
	}()

	select {
	case results := <-done:
		return results
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for files")
		return nil
	}
}
