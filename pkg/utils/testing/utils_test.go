package testing

import (
	"testing"

	"github.com/trezorg/lingualeo/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestInsertIntoSlice(t *testing.T) {
	slice := []utils.Value{"1", "4", "6"}
	newSlice := utils.InsertIntoSlice(slice, 1, "2")
	assert.Equal(t, newSlice, []utils.Value{"1", "2", "4", "6"})
	newSlice = utils.InsertIntoSlice(newSlice, 1, "2")
	assert.Equal(t, newSlice, []utils.Value{"1", "2", "2", "4", "6"})
}

func TestCheckEitherCommandIsAvailableOnSystem(t *testing.T) {
	assert.Equal(t, true, utils.IsCommandAvailable("bash -c 'oops'"))
	assert.Equal(t, true, utils.IsCommandAvailable("bash"))
	assert.Equal(t, false, utils.IsCommandAvailable("xxxxxxxxxxx"))
}
