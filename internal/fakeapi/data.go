package fakeapi

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trezorg/lingualeo/pkg/api"
)

var (
	SoundURL    = "http://audiocdn.lingualeo.com/v2/3/102085-631152000.mp3"
	u1, _       = url.Parse("http://contentcdn.lingualeo.com/uploads/picture/31064.png")
	u2, _       = url.Parse("http://contentcdn.lingualeo.com/uploads/picture/335521.png")
	u3, _       = url.Parse("http://contentcdn.lingualeo.com/uploads/picture/374830.png")
	u4, _       = url.Parse("http://contentcdn.lingualeo.com/uploads/picture/620779.png")
	PictureUrls = []*url.URL{
		u1,
		u2,
		u3,
		u4,
	}
	ResponseData = []byte(`{"error_msg":"","translate_source":"base","is_user":0,
	"word_forms":[{"word":"accommodation","type":"прил."}],
	"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/3589594.png",
	"translate":[
		{"id":2569250,"value":"жильё","votes":5703,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/31064.png"},
		{"id":2718711,"value":"проживание","votes":1589,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/335521.png"},
		{"id":185932,"value":"размещение","votes":880,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/374830.png"},
		{"id":2735899,"value":"помещение","votes":268,"is_user":0,"pic_url":"http:\/\/contentcdn.lingualeo.com\/uploads\/picture\/620779.png"}
	],
	"transcription":"əkəədˈeɪːʃən","word_id":102085,"word_top":0,
	"sound_url":"http:\/\/audiocdn.lingualeo.com\/v2\/3\/102085-631152000.mp3"}`)
	Expected   = []string{"жильё", "проживание", "размещение", "помещение"}
	SearchWord = "accommodation"
)

func CheckResult(t *testing.T, res api.Result, searchWord string, expected []string) {
	words := make([]string, 0, len(res.Translate))
	for _, tr := range res.Translate {
		words = append(words, tr.Value)
	}

	assert.Equalf(t, res.Word, searchWord, "Incorrect search word: %s", searchWord)
	assert.Len(t, res.Translate, 4, "Incorrect number of translated words: %d. Expected: %d", len(res.Translate), len(expected))
	assert.Equalf(t, words, expected, "Incorrect translated words order: %s. Expected: %s",
		strings.Join(expected, ", "),
		strings.Join(words, ", "),
	)
}
