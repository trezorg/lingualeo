package main

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

const (
	data = `{"error_msg":"","translate_source":"base","is_user":0,
	"word_forms":[{"word":"accomodation","type":"прил."}],
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

func TestParseResponseJson(t *testing.T) {
	searchWord := "accomodation"
	reader := ioutil.NopCloser(bytes.NewReader([]byte(data)))
	res := &lingualeoResult{Word: searchWord}
	expected := []string{"размещение", "жильё", "проживание", "помещение"}
	getJSON(reader, res)
	res.parseAndSortTranslate()
	if res.Word != searchWord {
		t.Errorf("Incorrect search word: %s", searchWord)
	}
	if len(res.Words) != 4 {
		t.Errorf("Incorrect number of translated words: %d. Expected: 4", len(res.Words))
	}
	if !reflect.DeepEqual(res.Words, expected) {
		t.Errorf(
			"Incorrect translated words order: %s. Expected: %s",
			strings.Join(expected, ", "),
			strings.Join(res.Words, ", "),
		)
	}
}
