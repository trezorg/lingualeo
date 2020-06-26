package fakeapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trezorg/lingualeo/pkg/api"
)

func CheckResult(t *testing.T, res api.Result, searchWord string, expected []string) {
	assert.Equalf(t, res.Word, searchWord, "Incorrect search word: %s", searchWord)
	assert.Len(t, res.Words, 4, "Incorrect number of translated words: %d. Expected: %d", len(res.Words), len(expected))
	assert.Equalf(t, res.Words, expected, "Incorrect translated words order: %s. Expected: %s",
		strings.Join(expected, ", "),
		strings.Join(res.Words, ", "),
	)
}
