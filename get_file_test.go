package main

import (
	"sync"
	"testing"
)

var (
	resData = resultFile{Filename: "/tmp/lingualeo.mp3"}
)

func TestGetWordFilePathResponse(t *testing.T) {
	var mockGetWordFilePathResponse = func(url string, idx int, wg *sync.WaitGroup) resultFile {
		return resData
	}
	origGetWordFilePathResponse := getWordFilePath
	getWordFilePath = mockGetWordFilePathResponse
	defer func() { getWordFilePath = origGetWordFilePathResponse }()

	out := downloadFiles("http://test.com/file")
	fileName := (<-out).(resultFile).Filename
	if fileName != resData.Filename {
		t.Errorf("Incorrect filename: %s", fileName)
	}
}
