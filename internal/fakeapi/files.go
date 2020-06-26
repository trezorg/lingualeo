package fakeapi

import (
	"bufio"
	"io"

	"github.com/trezorg/lingualeo/pkg/files"
)

var TestFile = "/tmp/test.file"

type testWriteCloser struct {
	*bufio.Writer
}

func (twc *testWriteCloser) Close() error {
	return nil
}

type fakeFileDownloader struct {
}

func (f *fakeFileDownloader) GetWriter() (io.WriteCloser, string, error) {
	var b *testWriteCloser
	return b, TestFile, nil
}

func (f *fakeFileDownloader) DownloadFile() (string, error) {
	_, fl, err := f.GetWriter()
	return fl, err
}

func FakeDownloader(string) files.Downloader {
	return &fakeFileDownloader{}
}
