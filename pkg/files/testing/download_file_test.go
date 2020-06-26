package testing

import (
	"context"
	"testing"

	"github.com/trezorg/lingualeo/internal/fakeapi"

	"github.com/trezorg/lingualeo/pkg/files"

	"github.com/stretchr/testify/assert"
)

func TestDownloadWordFile(t *testing.T) {

	inChan := make(chan string, 1)
	inChan <- "http://test.com/file"

	ctx := context.Background()

	close(inChan)

	out := files.DownloadFiles(ctx, inChan, fakeapi.FakeDownloader)
	fileName := (<-out).Filename
	assert.Equal(t, fileName, fakeapi.TestFile)
}
