package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

type convertibleBoolean bool

func (bit *convertibleBoolean) UnmarshalJSON(data []byte) error {
	asString := strings.Trim(string(data), "\"")
	if asString == "1" || asString == "true" {
		*bit = true
	} else if asString == "0" || asString == "false" {
		*bit = false
	} else {
		return fmt.Errorf("boolean unmarshal error: invalid input %s", asString)
	}
	return nil
}

func prepareClient() (*http.Client, error) {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		return nil, err
	}
	netTransport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Jar:       jar,
		Transport: netTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			if len(via) == 0 {
				return nil
			}
			for attr, val := range via[0].Header {
				if _, ok := req.Header[attr]; !ok {
					req.Header[attr] = val
				}
			}
			return nil
		},
	}
	return client, nil
}

func auth(args *lingualeoArgs, client *http.Client) error {
	values := map[string]string{
		"email":              args.Email,
		"password":           args.Password,
		"type":               "login",
		"successRedirectUrl": "",
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return nil
	}
	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonValue)))
	for key, values := range agentHeaders {
		for _, header := range values {
			req.Header.Add(key, header)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	body, err := readBody(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("response status code: %d\nbody:\n%s", resp.StatusCode, *body)
	}
	res := &responseError{}
	err = getJSONFromString(body, res)
	if err != nil {
		return err
	}
	if res.ErrorCode != 0 {
		return fmt.Errorf(res.ErrorMsg)
	}
	return nil
}

func getWordResponseBody(word string, client *http.Client) (*string, error) {
	req, err := http.NewRequest("POST", translateURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, values := range agentHeaders {
		for _, header := range values {
			req.Header.Add(key, header)
		}
	}
	q := req.URL.Query()
	q.Add("word", word)
	q.Add("include_media", "1")
	q.Add("app_word_forms", "1")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	body, err := readBody(resp)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"Response status code: %d\nword: %s\nbody:\n%s",
			resp.StatusCode,
			word,
			*body,
		)
	}
	return body, err
}

func getFileContent(url string, idx int, wg *sync.WaitGroup) resultFile {
	defer wg.Done()
	file, err := ioutil.TempFile("/tmp", "lingualeo")
	if err != nil {
		return resultFile{Error: err, Index: idx}
	}
	fd, err := os.Create(file.Name())
	if err != nil {
		return resultFile{Error: err, Index: idx}
	}
	defer func() {
		err := fd.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	resp, err := http.Get(url)
	if err != nil {
		return resultFile{Error: fmt.Errorf("cannot read sound url: %s, %w", url, err), Index: idx}
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return resultFile{Error: fmt.Errorf("bad status: %s", resp.Status), Index: idx}
	}
	_, err = io.Copy(fd, resp.Body)
	if err != nil {
		return resultFile{Error: err, Index: idx}
	}
	return resultFile{Filename: file.Name(), Index: idx}
}

var getWordResponseString = getWordResponseBody
var getWordFilePath = getFileContent

func getWord(word string, client *http.Client, out chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	body, err := getWordResponseString(word, client)
	if err != nil {
		out <- translateResult{Error: err}
		return
	}
	res := &lingualeoResult{Word: word}
	err = getJSONFromString(body, res)
	if err != nil {
		res := &lingualeoNoResult{}
		if getJSONFromString(body, res) == nil {
			out <- translateResult{Error: fmt.Errorf("cannot translate word: %s", word)}
			return
		}
		out <- translateResult{Error: err}
		return
	}
	if len(res.ErrorMsg) > 0 {
		out <- translateResult{Error: fmt.Errorf(res.ErrorMsg)}
		return
	}
	res.parseAndSortTranslate()
	out <- translateResult{Result: res}
}

func addWord(res lingualeoResult, client *http.Client, out chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	values := map[string]string{
		"word":  res.Word,
		"tword": strings.Join(res.Words, ", "),
		"port":  "1001",
	}
	jsonValue, _ := json.Marshal(values)
	req, err := http.NewRequest("POST", addWordURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		out <- translateResult{Error: err}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonValue)))
	for key, values := range agentHeaders {
		for _, header := range values {
			req.Header.Add(key, header)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		out <- translateResult{Error: err}
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	body, err := readBody(resp)
	if err != nil {
		out <- translateResult{Error: err}
		return
	}
	if resp.StatusCode != 200 {
		out <- translateResult{Error: fmt.Errorf(
			"Response status code: %d\nword: %s\nbody:\n%s",
			resp.StatusCode,
			res.Word,
			*body,
		)}
		return
	}
	lingRes := &lingualeoResult{Word: res.Word}
	err = getJSONFromString(body, lingRes)
	if err != nil {
		out <- translateResult{Error: err}
		return
	}
	if len(lingRes.ErrorMsg) > 0 {
		out <- translateResult{Error: fmt.Errorf(res.ErrorMsg)}
		return
	}
	out <- translateResult{Result: &res}
}

func getWords(words []string, client *http.Client) <-chan interface{} {
	count := len(words)
	out := make(chan interface{}, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for _, word := range words {
		go getWord(word, client, out, &wg)
	}
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func addWords(results []lingualeoResult, client *http.Client) <-chan interface{} {
	count := len(results)
	out := make(chan interface{}, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for _, res := range results {
		go addWord(res, client, out, &wg)
	}
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func downloadFiles(urls ...string) <-chan interface{} {
	count := len(urls)
	out := make(chan interface{}, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for idx, url := range urls {
		out <- getWordFilePath(url, idx, &wg)
	}
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}
