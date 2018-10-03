package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/wsxiaoys/terminal/color"
)

func unique(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func parseAndSortTranslate(result *lingualeoResult) {
	sort.Slice(result.Translate, func(i, j int) bool {
		return result.Translate[i].Votes > result.Translate[j].Votes
	})
	for _, translate := range result.Translate {
		for _, word := range sanitizeWords(&translate) {
			result.Words = append(result.Words, word)
		}
	}
	result.Words = unique(result.Words)
}

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

func printColorString(clr string, text string) {
	str := fmt.Sprintf("@{%s}%s\n", clr, text)
	color.Printf(str)
}
