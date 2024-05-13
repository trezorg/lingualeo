package translator

import (
	"unicode"
)

func isRussianWord(s string) bool {
	for _, symbol := range s {
		if !unicode.Is(unicode.Cyrillic, symbol) && !unicode.Is(unicode.Number, symbol) {
			return false
		}
	}
	return true
}

func isEnglishWord(s string) bool {
	for _, symbol := range s {
		if !unicode.Is(unicode.Latin, symbol) && !unicode.Is(unicode.Number, symbol) {
			return false
		}
	}
	return true
}
