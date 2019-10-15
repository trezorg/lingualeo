package main

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	resData = resultFile{Filename: "/tmp/lingualeo.mp3"}
)

func TestGetWordFilePathResponse(t *testing.T) {
	var mockGetWordFilePathResponse = func(url string, idx int, out chan<- interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		out <- resData
	}
	origGetWordFilePathResponse := getWordFilePath
	getWordFilePath = mockGetWordFilePathResponse
	defer func() { getWordFilePath = origGetWordFilePathResponse }()

	inChan := make(chan interface{}, 1)
	inChan <- "http://test.com/file"

	ctx := context.Background()

	close(inChan)

	out := downloadFiles(ctx, inChan)
	fileName := (<-out).(resultFile).Filename
	assert.Equal(t, fileName, resData.Filename)
}
