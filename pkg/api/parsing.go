package api

import (
	"fmt"
	"regexp"
	"strings"
)

func fixTranslateString(word string) string {
	word = blankSymbolsRegex.ReplaceAllString(word, " ")
	word = blankSymbolsWithPointRegex.ReplaceAllString(word, ". ")
	return strings.ToLower(word)
}

func removeSymbols(word string) string {
	nonAlphaBet := regexp.MustCompile(fmt.Sprintf(`[^%s]`, alphabet))
	return nonAlphaBet.ReplaceAllString(word, "")
}

func sanitizeWords(word string) []string {
	var results []string
	words := splitRegex.Split(strings.TrimSpace(word), -1)
	for _, word := range words {
		word = fixTranslateString(strings.TrimSpace(removeSymbols(word)))
		if len(word) > 1 {
			results = append(results, word)
		}
	}
	return results
}
