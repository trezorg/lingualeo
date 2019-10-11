package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertIntoSlice(t *testing.T) {
	slice := []Value{"1", "4", "6"}
	newSlice := insertIntoSlice(slice, 1, "2")
	assert.Equal(t, newSlice, []Value{"1", "2", "4", "6"})
	newSlice = insertIntoSlice(newSlice, 1, "2")
	assert.Equal(t, newSlice, []Value{"1", "2", "2", "4", "6"})
}
