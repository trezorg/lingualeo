package translator

import (
	"log/slog"
	"testing"

	"github.com/trezorg/lingualeo/internal/api"
	"github.com/trezorg/lingualeo/internal/fakeapi"
	"github.com/trezorg/lingualeo/internal/logger"
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
	player := NewMock_Pronouncer(t)
	outputer := NewMock_Outputer(t)
	res := translateWordResult(fakeapi.SearchWord)

	downloader.EXPECT().Download(fakeapi.SoundURL).Return(testFile, nil).Times(count)
	downloader.EXPECT().Remove(testFile).Return(nil).Times(count)
	translator.EXPECT().TranslateWord(fakeapi.SearchWord).Return(res).Times(count)
	player.EXPECT().Play(t.Context(), testFile).Return(nil).Times(count)

	logger.Prepare(slog.LevelError + 10)
	searchWords := make([]string, 0, count)

	for range count {
		searchWords = append(searchWords, fakeapi.SearchWord)
	}
	args := Lingualeo{
		Sound:             true,
		Words:             searchWords,
		Add:               false,
		Visualise:         true,
		DownloadSoundFile: true,
		Translator:        translator,
		Downloader:        downloader,
		Pronouncer:        player,
		Outputer:          outputer,
	}

	ch := args.translateToChan(t.Context())

	for result := range ch {
		fakeapi.CheckResult(t, result, searchWords[0], fakeapi.Expected)
	}
}
