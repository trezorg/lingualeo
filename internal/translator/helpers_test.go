package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckEitherCommandIsAvailableOnSystem(t *testing.T) {
	assert.True(t, isCommandAvailable("bash -c 'oops'"))
	assert.True(t, isCommandAvailable("bash"))
	assert.False(t, isCommandAvailable("xxxxxxxxxxx"))
}

func TestRussianWord(t *testing.T) {
	assert.True(t, isRussianWord("гном"))
	assert.True(t, isRussianWord("гном1"))
	assert.False(t, isRussianWord("test"))
}
