package api

import (
	"net/http"
)

const (
	authURL      = "https://lingualeo.com/api/auth"
	translateURL = "https://api.lingualeo.com/getTranslates"
	addWordURL   = "https://api.lingualeo.com/addWord"
	apiVersion   = "1.0.1"
)

var (
	agentHeaders = http.Header{
		"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.116 Safari/537.36"},
		"Accept":     []string{"application/json"},
	}
)
