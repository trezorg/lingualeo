package main

import (
	"fmt"
	"regexp"
	"strings"
)

func fixTranslateString(word string) string {
	word = string(blankSymbolsRegex.ReplaceAllString(word, " "))
	word = blankSymbolsWithPointRegex.ReplaceAllString(word, ". ")
	return strings.ToLower(word)
}

func removeSymbols(word string) string {
	nonAphaBet := regexp.MustCompile(fmt.Sprintf(`[^%s]`, alphabet))
	return string(nonAphaBet.ReplaceAllString(word, ""))
}

func sanitizeWords(result *lingualeoWordResult) []string {
	var results []string
	words := splitRegex.Split(strings.TrimSpace(result.Value), -1)
	for _, word := range words {
		word = fixTranslateString(strings.TrimSpace(removeSymbols(word)))
		if len(word) > 1 {
			results = append(results, word)
		}
	}
	return results
}
