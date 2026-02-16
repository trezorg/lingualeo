package translator

import (
	"context"
	"sync"

	"github.com/trezorg/lingualeo/internal/channel"
	"github.com/trezorg/lingualeo/internal/files"
)

// Downloader interface
//
//go:generate mockery
type Downloader interface {
	Download(ctx context.Context, url string) (string, error)
	Remove(path string) error
}

// downloadFiles download files from URLs channel
func downloadFiles(ctx context.Context, urls <-chan string, downloader Downloader) <-chan files.File {
	out := make(chan files.File)
	var wg sync.WaitGroup
	wg.Go(func() {
		idx := 0
		for url := range channel.OrDone(ctx, urls) {
			idxCopy := idx
			wg.Go(func() {
				filename, err := downloader.Download(ctx, url)
				out <- files.File{Error: err, Filename: filename, Index: idxCopy}
			})
			idx++
		}
	})
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
