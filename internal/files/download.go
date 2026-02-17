package files

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/trezorg/lingualeo/internal/validator"
)

const (
	fileTemplate       = "lingualeo"
	defaultHTTPTimeout = 30 * time.Second
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
type FileDownloader struct {
	client *http.Client
}

// New initialize new file downloader
func New() *FileDownloader {
	return &FileDownloader{
		client: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

// Writer prepares WriteCloser for temporary file
func (*FileDownloader) Writer() (io.WriteCloser, string, error) {
	fd, err := os.CreateTemp(os.TempDir(), fileTemplate)
	if err != nil {
		return nil, "", err
	}
	return fd, fd.Name(), nil
}

// Download downloads file
func (f *FileDownloader) Download(ctx context.Context, url string) (string, error) {
	if err := validator.ValidateURL(url); err != nil {
		return "", fmt.Errorf("invalid download URL: %w", err)
	}
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("cannot read URL: %s, %w", url, err)
	}
	resp, err := f.client.Do(req) //nolint:gosec // URL is validated before request execution
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

func (*FileDownloader) Remove(path string) error {
	return os.Remove(path)
}

// DownloadBytes downloads file into bytes slice
func (f *FileDownloader) DownloadBytes(ctx context.Context, url string) ([]byte, error) {
	if err := validator.ValidateURL(url); err != nil {
		return nil, fmt.Errorf("invalid download URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot read URL: %s, %w", url, err)
	}
	resp, err := f.client.Do(req) //nolint:gosec // URL is validated before request execution
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
