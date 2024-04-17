package fakeapi

import (
	"bufio"
	"io"
)

var TestFile = "/tmp/test.file"

type testWriteCloser struct {
	*bufio.Writer
}

func (twc *testWriteCloser) Close() error {
	return nil
}

type FakeFileDownloader struct {
}

func (f *FakeFileDownloader) Writer() (io.WriteCloser, string, error) {
	var b *testWriteCloser
	return b, TestFile, nil
}

func (f *FakeFileDownloader) Download(_ string) (string, error) {
	_, fl, err := f.Writer()
	return fl, err
}
