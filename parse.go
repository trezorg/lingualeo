package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wsxiaoys/terminal/color"
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
		if len(word) > 0 {
			results = append(results, word)
		}
	}
	return results
}

func printTranslate(result *lingualeoResult) {
	var strTitle string
	if result.Exists {
		strTitle = "existing"
	} else {
		strTitle = "new"
	}
	color.Printf("@{r}Found %s word:\n", strTitle)
	color.Printf("@{g}['%s'] (%s)\n", result.Word, result.Transcription)
	for _, word := range result.Words {
		printColorString("y", word)
	}
}

func printAddTranslate(result *lingualeoResult) {
	var strTitle string
	if result.Exists {
		strTitle = "Updated existing"
	} else {
		strTitle = "Added new"
	}
	color.Printf("@{r}%s word: ", strTitle)
	color.Printf("@{g}['%s']\n", result.Word)
}
