package translator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadWordFile(t *testing.T) {
	downloader := NewMock_Downloader(t)
	url := "http://test.com/file"
	testFile := "/tmp/test.file"

	downloader.EXPECT().Download(url).Return(testFile, nil).Once()

	inChan := make(chan string, 1)
	inChan <- url

	ctx := context.Background()

	close(inChan)

	out := DownloadFiles(ctx, inChan, downloader)
	fileName := (<-out).Filename
	assert.Equal(t, fileName, testFile)
}
