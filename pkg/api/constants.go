package api

import (
	"net/http"
	"regexp"
	"strings"
)

const (
	authURL            = "https://lingualeo.com/api/login"
	translateURL       = "https://api.lingualeo.com/gettranslates?port=1001"
	addWordURL         = "https://api.lingualeo.com/addword?port=1001"
	bigRussianAlphabet = "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ"
	apiVersion         = "1.0.1"
)

var (
	agentHeaders = http.Header{
		"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.116 Safari/537.36"},
		"Accept":     []string{"application/json"},
	}
	splitRegex                 = regexp.MustCompile(`\s*?[:,;]+\s*?`)
	blankSymbolsRegex          = regexp.MustCompile(`\s+`)
	blankSymbolsWithPointRegex = regexp.MustCompile(`\s+\.\s*`)
	alphabet                   = strings.Join(
		[]string{symbols, bigRussianAlphabet, strings.ToLower(bigRussianAlphabet)},
		"",
	)
	symbols = "-. "
)
