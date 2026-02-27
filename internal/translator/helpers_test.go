package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckEitherCommandIsAvailableOnSystem(t *testing.T) {
	assert.True(t, isCommandAvailable("bash -c 'oops'"))
	assert.True(t, isCommandAvailable("bash -c \"echo test\""))
	assert.True(t, isCommandAvailable("bash"))
	assert.False(t, isCommandAvailable("xxxxxxxxxxx"))
	assert.False(t, isCommandAvailable("bash -c \"oops"))
}

func TestRussianWord(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected bool
	}{
		// Basic Russian words
		{name: "simple russian word", word: "гном", expected: true},
		{name: "russian word with number", word: "гном1", expected: true},
		{name: "russian word with multiple numbers", word: "слово123", expected: true},
		{name: "only numbers", word: "12345", expected: true},
		{name: "russian uppercase", word: "МОСКВА", expected: true},
		{name: "russian mixed case", word: "МосквА", expected: true},

		// English words
		{name: "simple english word", word: "test", expected: false},
		{name: "english word with number", word: "test1", expected: false},
		{name: "english uppercase", word: "HELLO", expected: false},

		// Mixed Cyrillic/Latin
		{name: "mixed cyrillic latin", word: "testгном", expected: false},
		{name: "mixed latin cyrillic", word: "гномtest", expected: false},

		// Special characters
		{name: "with hyphen", word: "как-то", expected: false},
		{name: "with space", word: "привет мир", expected: false},
		{name: "with punctuation", word: "привет!", expected: false},
		{name: "with underscore", word: "при_вет", expected: false},

		// Empty and edge cases
		{name: "empty string", word: "", expected: true}, // Empty passes because loop doesn't run
		{name: "single cyrillic char", word: "а", expected: true},
		{name: "single number", word: "1", expected: true},

		// Unicode edge cases
		{name: "russian yo", word: "ёлка", expected: true},
		{name: "russian hard sign", word: "съезд", expected: true},
		{name: "russian soft sign", word: "день", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isRussianWord(tt.word))
		})
	}
}

func TestEnglishWord(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected bool
	}{
		// Basic English words
		{name: "simple english word", word: "hello", expected: true},
		{name: "english word with number", word: "test1", expected: true},
		{name: "english word with multiple numbers", word: "word123", expected: true},
		{name: "only numbers", word: "12345", expected: true},
		{name: "english uppercase", word: "HELLO", expected: true},
		{name: "english mixed case", word: "HeLLo", expected: true},

		// Russian words
		{name: "simple russian word", word: "гном", expected: false},
		{name: "russian word with number", word: "гном1", expected: false},
		{name: "russian uppercase", word: "МОСКВА", expected: false},

		// Mixed Cyrillic/Latin
		{name: "mixed cyrillic latin", word: "testгном", expected: false},
		{name: "mixed latin cyrillic", word: "гномtest", expected: false},

		// Special characters
		{name: "with hyphen", word: "some-thing", expected: false},
		{name: "with space", word: "hello world", expected: false},
		{name: "with punctuation", word: "hello!", expected: false},
		{name: "with underscore", word: "hello_world", expected: false},

		// Empty and edge cases
		{name: "empty string", word: "", expected: true}, // Empty passes because loop doesn't run
		{name: "single latin char", word: "a", expected: true},
		{name: "single number", word: "1", expected: true},

		// Unicode edge cases - these are Latin script variants
		{name: "german umlaut", word: "größe", expected: true},
		{name: "french accent", word: "café", expected: true},
		{name: "spanish tilde", word: "año", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isEnglishWord(tt.word))
		})
	}
}

func TestCheckArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    Lingualeo
		wantErr bool
	}{
		{
			name: "valid args",
			args: Lingualeo{
				Email:    "user@example.com",
				Password: "password",
				Words:    []string{"test"},
			},
			wantErr: false,
		},
		{
			name: "missing email",
			args: Lingualeo{
				Password: "password",
				Words:    []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "invalid email - no @",
			args: Lingualeo{
				Email:    "userexample.com",
				Password: "password",
				Words:    []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "invalid email - no domain",
			args: Lingualeo{
				Email:    "user@",
				Password: "password",
				Words:    []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "missing password",
			args: Lingualeo{
				Email: "user@example.com",
				Words: []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "missing words",
			args: Lingualeo{
				Email:    "user@example.com",
				Password: "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.checkArgs()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
