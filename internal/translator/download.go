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

type downloadJob struct {
	url   string
	index int
}

func downloadFiles(ctx context.Context, urls <-chan string, downloader Downloader, workers int) <-chan files.File {
	out := make(chan files.File)
	jobs := make(chan downloadJob)
	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(jobs)
		idx := 0
		for url := range channel.OrDone(ctx, urls) {
			if !sendToChanWithContext(ctx, jobs, downloadJob{url: url, index: idx}) {
				return
			}
			idx++
		}
	})
	for range workerCount(workers) {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}
					filename, err := downloader.Download(ctx, job.url)
					if !sendToChanWithContext(ctx, out, files.File{Error: err, Filename: filename, Index: job.index}) {
						return
					}
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
