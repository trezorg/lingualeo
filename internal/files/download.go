package files

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

const (
	fileTemplate = "lingualeo"
	filePath     = "/tmp"
)

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
type FileDownloader struct{}

// New initialize new file downloader
func New() *FileDownloader {
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
			slog.Error("cannot close write file descriptor", "error", cErr)
		}
	}()
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot read URL: %s, %w", url, err)
	}
	defer func() {
		cErr := resp.Body.Close()
		if cErr != nil {
			slog.Error("cannot close response body", "error", cErr)
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

func (f *FileDownloader) Remove(path string) error {
	return os.Remove(path)
}

// DownloadBytes downloads file into bytes slice
func (f *FileDownloader) DownloadBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("cannot read URL: %s, %w", url, err)
	}
	defer func() {
		cErr := resp.Body.Close()
		if cErr != nil {
			slog.Error("cannot close response body", "error", cErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}
