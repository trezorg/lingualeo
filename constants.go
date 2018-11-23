package main

import (
	"net/http"
	"regexp"
	"strings"
)

const (
	bigRussianAlphabet = "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ"
	authURL            = "http://api.lingualeo.com/api/login"
	translateURL       = "http://api.lingualeo.com/gettranslates"
	addWordURL         = "http://api.lingualeo.com/addword"
)

var (
	defaultConfigFiles = []string{
		"lingualeo.conf",
		"lingualeo.yml",
		"lingualeo.yaml",
		"lingualeo.json",
	}
	splitRegex                 = regexp.MustCompile(`\s*?[:,;]+\s*?`)
	blankSymbolsRegex          = regexp.MustCompile(`\s+`)
	blankSymbolsWithPointRegex = regexp.MustCompile(`\s+\.\s*`)
	alphabet                   = strings.Join(
		[]string{symbols, bigRussianAlphabet, strings.ToLower(bigRussianAlphabet)},
		"",
	)
	symbols      = "-. "
	agentHeaders = http.Header{
		"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.116 Safari/537.36"},
		"Accept":     []string{"application/json"},
	}
)
