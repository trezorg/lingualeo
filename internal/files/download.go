package files

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/trezorg/lingualeo/internal/logger"
	"github.com/trezorg/lingualeo/pkg/channel"
)

const fileTemplate = "lingualeo"
const filePath = "/tmp"

// Downloader interface
type Downloader interface {
	Download(url string) (string, error)
	Writer() (io.WriteCloser, string, error)
}

// NewDownloader function type
type NewDownloader func(url string) Downloader

// File represents file for downloading
type File struct {
	Error    error
	Filename string
	Index    int
}

// GetIndex returns index from file structure
func (f File) GetIndex() int {
	return f.Index
}

// FileDownloader structure
type FileDownloader struct {
}

// NewFileDownloader initialize new file downloader
func NewFileDownloader() *FileDownloader {
	return &FileDownloader{}
}

// Writer prepares WriteCloser for temporary file
func (f *FileDownloader) Writer() (io.WriteCloser, string, error) {
	fl, err := os.CreateTemp(filePath, fileTemplate)
	if err != nil {
		return nil, "", err
	}
	fd, err := os.Create(fl.Name())
	if err != nil {
		return nil, "", err
	}
	return fd, fl.Name(), nil
}

// Download downloads file
func (f *FileDownloader) Download(url string) (string, error) {
	fd, filename, err := f.Writer()
	if err != nil {
		return "", err
	}
	defer func() {
		cErr := fd.Close()
		if cErr != nil {
			logger.Error(cErr)
		}
	}()
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot read URL: %s, %w", url, err)
	}
	defer func() {
		cErr := resp.Body.Close()
		if cErr != nil {
			logger.Error(cErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}
	_, err = io.Copy(fd, resp.Body)
	if err != nil {
		return "", err
	}
	return filename, nil
}

// DownloadFiles download files from URLs channel
func DownloadFiles(ctx context.Context, urls <-chan string, downloader Downloader) <-chan File {
	out := make(chan File)
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
				out <- File{Error: err, Filename: filename, Index: idx}
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
