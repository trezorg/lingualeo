package translator

import (
	"context"
	"sync"
	"testing"

	"github.com/trezorg/lingualeo/internal/logger"

	"github.com/trezorg/lingualeo/internal/fakeapi"
)

func TestProcessTranslationResponseJson(t *testing.T) {

	count := 1000 // max for race checking
	logger.InitLogger("FATAL", true)
	searchWords := make([]string, 0, count)

	for i := 0; i < count; i++ {
		searchWords = append(searchWords, fakeapi.SearchWord)
	}
	ctx := context.Background()
	fakeAPI := fakeapi.FakeAPI{}

	args := Lingualeo{Sound: true, Words: searchWords, Add: false, API: &fakeAPI}
	var wg sync.WaitGroup

	wg.Add(1)
	soundChan, _, resultChan := args.Process(ctx, &wg)
	wg.Add(1)

	go args.downloadAndPronounce(ctx, soundChan, &wg, &fakeapi.FakeFileDownloader{})

	for result := range resultChan {
		fakeapi.CheckResult(t, result, searchWords[0], fakeapi.Expected)
	}

	wg.Wait()

}
