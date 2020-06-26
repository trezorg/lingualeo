package files

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/trezorg/lingualeo/pkg/channel"
	"github.com/trezorg/lingualeo/pkg/logger"
)

const fileTemplate = "lingualeo"
const filePath = "/tmp"

type Downloader interface {
	DownloadFile() (string, error)
	GetWriter() (io.WriteCloser, string, error)
}

type NewDownloader func(url string) Downloader

type File struct {
	Error    error
	Filename string
	Index    int
}

func (f File) GetIndex() int {
	return f.Index
}

type FileDownloader struct {
	URL string
}

func NewFileDownloader(url string) Downloader {
	return &FileDownloader{URL: url}
}

func (f *FileDownloader) GetWriter() (io.WriteCloser, string, error) {
	fl, err := ioutil.TempFile(filePath, fileTemplate)
	if err != nil {
		return nil, "", err
	}
	fd, err := os.Create(fl.Name())
	if err != nil {
		return nil, "", err
	}
	return fd, fl.Name(), nil
}

func (f *FileDownloader) DownloadFile() (string, error) {
	fd, filename, err := f.GetWriter()
	if err != nil {
		return "", err
	}
	defer func() {
		err := fd.Close()
		if err != nil {
			logger.Log.Error(err)
		}
	}()
	resp, err := http.Get(f.URL)
	if err != nil {
		return "", fmt.Errorf("cannot read url: %s, %w", f.URL, err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Log.Error(err)
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

func DownloadFiles(ctx context.Context, urls <-chan string, downloader NewDownloader) <-chan File {
	out := make(chan File)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		idx := 0
		for url := range channel.OrStringDone(ctx, urls) {
			wg.Add(1)
			go func(idx int, url string) {
				defer wg.Done()
				filename, err := downloader(url).DownloadFile()
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
