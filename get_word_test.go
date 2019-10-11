package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	responseData = `{"error_msg":"","translate_source":"base","is_user":0,
	"word_forms":[{"word":"accommodation","type":"прил."}],
	"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/3589594.png",
	"translate":[
		{"id":33404925,"value":"размещение; жильё","votes":6261,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/3589594.png"},
		{"id":2569250,"value":"жильё","votes":5703,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/31064.png"},
		{"id":2718711,"value":"проживание","votes":1589,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/335521.png"},
		{"id":185932,"value":"размещение","votes":880,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/374830.png"},
		{"id":2735899,"value":"помещение","votes":268,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/620779.png"}
	],
	"transcription":"əkəədˈeɪːʃən","word_id":102085,"word_top":0,
	"sound_url":"http:\/\/audiocdn.lingualeo.com\/v2\/3\/102085-631152000.mp3"}`
)

func checkResult(t *testing.T, res *lingualeoResult, searchWord string, expected []string) {
	assert.Equalf(t, res.Word, searchWord, "Incorrect search word: %s", searchWord)
	assert.Len(t, res.Words, 4, "Incorrect number of translated words: %d. Expected: %d", len(res.Words), len(expected))
	assert.Equalf(t, res.Words, expected, "Incorrect translated words order: %s. Expected: %s",
		strings.Join(expected, ", "),
		strings.Join(res.Words, ", "),
	)
}

func TestParseResponseJson(t *testing.T) {
	searchWord := "accommodation"
	reader := ioutil.NopCloser(bytes.NewReader([]byte(responseData)))
	res := &lingualeoResult{Word: searchWord}
	expected := []string{"размещение", "жильё", "проживание", "помещение"}
	err := res.fillObjectFromJSON(reader)
	assert.NoError(t, err, "Cannot fill object from json")
	res.parseAndSortTranslate()
	checkResult(t, res, searchWord, expected)
}

func TestGetWordResponseJson(t *testing.T) {
	var mockGetWordResponseString = func(word string, client *http.Client) (*string, error) {
		return &responseData, nil
	}
	origGetWordResponseString := getWordResponseString
	getWordResponseString = mockGetWordResponseString
	defer func() { getWordResponseString = origGetWordResponseString }()

	searchWord := "accommodation"
	expected := []string{"размещение", "жильё", "проживание", "помещение"}

	out := make(chan interface{})
	defer close(out)
	var wg sync.WaitGroup
	client := &http.Client{}

	wg.Add(1)
	go getWord(searchWord, client, out, &wg)

	res := (<-out).(translateResult).Result
	checkResult(t, res, searchWord, expected)
}

func TestGetWordsResponseJson(t *testing.T) {
	var mockGetWordResponseString = func(word string, client *http.Client) (*string, error) {
		return &responseData, nil
	}
	origGetWordResponseString := getWordResponseString
	getWordResponseString = mockGetWordResponseString
	defer func() { getWordResponseString = origGetWordResponseString }()

	searchWords := []string{"accommodation"}
	expected := []string{"размещение", "жильё", "проживание", "помещение"}

	client := &http.Client{}

	out := getWords(searchWords, client)

	res := (<-out).(translateResult).Result
	checkResult(t, res, searchWords[0], expected)
}
