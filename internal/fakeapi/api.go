package fakeapi

import (
	"context"
	"sync"

	"github.com/trezorg/lingualeo/pkg/channel"

	"github.com/trezorg/lingualeo/pkg/api"
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
	Expected   = []string{"размещение", "жильё", "проживание", "помещение"}
	SearchWord = "accommodation"
)

type FakeAPI struct {
	*api.API
}

func (f *FakeAPI) TranslateWord(word string) api.OpResult {
	res := api.EnglishResult{Word: word}
	err := res.FromResponse(responseData)
	return api.OpResult{Result: &res, Error: err}
}

func (f *FakeAPI) AddWord(word string, _ []string) api.OpResult {
	res := api.EnglishResult{Word: word}
	err := res.FromResponse(responseData)
	return api.OpResult{Result: &res, Error: err}
}

func (f *FakeAPI) TranslateWords(ctx context.Context, results <-chan string) <-chan api.OpResult {
	out := make(chan api.OpResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for word := range channel.OrStringDone(ctx, results) {
			wg.Add(1)
			go func(word string) {
				defer wg.Done()
				out <- f.TranslateWord(word)
			}(word)
		}
	}()
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func (f *FakeAPI) AddWords(ctx context.Context, results <-chan api.Result) <-chan api.OpResult {
	out := make(chan api.OpResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for res := range api.OrResultDone(ctx, results) {
			wg.Add(1)
			result := res
			go func(result api.Result) {
				defer wg.Done()
				out <- f.AddWord(result.GetWord(), result.GetTranslate())
			}(result)
		}
	}()
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}
