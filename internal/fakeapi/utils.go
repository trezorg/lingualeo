package fakeapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trezorg/lingualeo/pkg/api"
)

func CheckResult(t *testing.T, res api.Result, searchWord string, expected []string) {
	assert.Equalf(t, res.GetWord(), searchWord, "Incorrect search word: %s", searchWord)
	assert.Len(t, res.GetTranslate(), 4, "Incorrect number of translated words: %d. Expected: %d", len(res.GetTranslate()), len(expected))
	assert.Equalf(t, res.GetTranslate(), expected, "Incorrect translated words order: %s. Expected: %s",
		strings.Join(expected, ", "),
		strings.Join(res.GetTranslate(), ", "),
	)
}
