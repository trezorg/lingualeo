package translator

import (
	"context"
	"sync"

	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/pkg/channel"
)

// Downloader interface
//
//go:generate mockery
type Downloader interface {
	Download(url string) (string, error)
	Remove(path string) error
}

// downloadFiles download files from URLs channel
func downloadFiles(ctx context.Context, urls <-chan string, downloader Downloader) <-chan files.File {
	out := make(chan files.File)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		idx := 0
		for url := range channel.OrDone(ctx, urls) {
			wg.Add(1)
			go func(idx int, url string) {
				defer wg.Done()
				filename, err := downloader.Download(url)
				out <- files.File{Error: err, Filename: filename, Index: idx}
			}(idx, url)
			idx++
		}
	}()
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}
