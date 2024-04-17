package testing

import (
	"context"
	"testing"

	"github.com/trezorg/lingualeo/internal/fakeapi"
	"github.com/trezorg/lingualeo/pkg/channel"

	"github.com/stretchr/testify/assert"
)

func TestTranslateWord(t *testing.T) {
	res := (&fakeapi.FakeAPI{}).TranslateWord(fakeapi.SearchWord)
	assert.NoError(t, res.Error, "Cannot get object from json")
	fakeapi.CheckResult(t, res.Result, fakeapi.SearchWord, fakeapi.Expected)
}

func TestTranslateWords(t *testing.T) {
	searchWords := []string{fakeapi.SearchWord}
	fakeAPI := fakeapi.FakeAPI{}
	ctx := context.Background()
	ch := channel.ToChannel(ctx, searchWords...)
	out := fakeAPI.TranslateWords(ctx, ch)
	res := (<-out).Result
	fakeapi.CheckResult(t, res, searchWords[0], fakeapi.Expected)
}
