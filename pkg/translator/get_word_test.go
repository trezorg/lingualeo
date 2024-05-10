package translator

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/trezorg/lingualeo/internal/logger"
	"github.com/trezorg/lingualeo/pkg/api"

	"github.com/trezorg/lingualeo/internal/fakeapi"
)

func translateWordResult(word string) api.OperationResult {
	res := api.Result{Word: word}
	err := res.FromResponse(fakeapi.ResponseData)
	return api.OperationResult{Result: res, Error: err}
}

func TestProcessTranslationResponseJson(t *testing.T) {
	downloader := NewMock_Downloader(t)
	testFile := "/tmp/test.file"
	count := 1000 // max for race checking
	translator := NewMock_Translator(t)

	res := translateWordResult(fakeapi.SearchWord)

	downloader.EXPECT().Download(fakeapi.SoundURL).Return(testFile, nil).Times(count)
	translator.EXPECT().TranslateWord(fakeapi.SearchWord).Return(res).Times(count)

	logger.Prepare(slog.LevelError + 10)
	searchWords := make([]string, 0, count)

	for i := 0; i < count; i++ {
		searchWords = append(searchWords, fakeapi.SearchWord)
	}
	ctx := context.Background()

	args := Lingualeo{Sound: true, Words: searchWords, Add: false, Translator: translator}
	var wg sync.WaitGroup

	wg.Add(1)
	soundChan, _, resultChan := args.Process(ctx, &wg)
	wg.Add(1)

	go args.downloadAndPronounce(ctx, soundChan, &wg, downloader)

	for result := range resultChan {
		fakeapi.CheckResult(t, result, searchWords[0], fakeapi.Expected)
	}

	wg.Wait()
}
