package main

import (
	"bytes"
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

type formValue struct {
	Name  string
	Value string
}

type convertibleBoolean bool

func (bit *convertibleBoolean) UnmarshalJSON(data []byte) error {
	asString := string(data)
	if asString == "1" || asString == "true" {
		*bit = true
	} else if asString == "0" || asString == "false" {
		*bit = false
	} else {
		return fmt.Errorf("Boolean unmarshal error: invalid input %s", asString)
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
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("Too many redirects")
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

func readBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getJSON(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func auth(args *lingualeoArgs, client *http.Client) error {
	values := map[string]string{
		"email":    args.Email,
		"password": args.Password,
	}
	jsonValue, _ := json.Marshal(values)
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
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := readBody(resp)
		return fmt.Errorf("Response status code: %d\nbody:\n%s", resp.StatusCode, body)
	}
	return nil
}

func getWord(word string, client *http.Client, out chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	req, err := http.NewRequest("POST", translateURL, nil)
	if err != nil {
		out <- result{Error: err}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	for key, values := range agentHeaders {
		for _, header := range values {
			req.Header.Add(key, header)
		}
	}
	q := req.URL.Query()
	q.Add("word", word)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		out <- result{Error: err}
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := readBody(resp)
		if err != nil {
			out <- result{Error: fmt.Errorf(
				"Response status code: %d\nword: %s\nbody:\n%s",
				resp.StatusCode,
				word,
				body,
			)}
			return
		}
	}
	res := &lingualeoResult{Word: word}
	err = getJSON(resp, res)
	if err != nil {
		out <- result{Error: err}
		return
	}
	if len(res.ErrorMsg) > 0 {
		out <- result{Error: fmt.Errorf(res.ErrorMsg)}
		return
	}
	out <- result{Result: res}
}

func addWord(res lingualeoResult, client *http.Client, out chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	values := map[string]string{
		"word":  res.Word,
		"tword": strings.Join(res.Words, ", "),
	}
	jsonValue, _ := json.Marshal(values)
	req, err := http.NewRequest("POST", addWordURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		out <- result{Error: err}
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
		out <- result{Error: err}
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := readBody(resp)
		if err != nil {
			out <- result{Error: fmt.Errorf(
				"Response status code: %d\nword: %s\nbody:\n%s",
				resp.StatusCode,
				res.Word,
				body,
			)}
			return
		}
	}
	lingRes := &lingualeoResult{Word: res.Word}
	err = getJSON(resp, lingRes)
	if err != nil {
		out <- result{Error: err}
		return
	}
	if len(lingRes.ErrorMsg) > 0 {
		out <- result{Error: fmt.Errorf(res.ErrorMsg)}
		return
	}
	out <- result{Result: &res}
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
		go func(url string, idx int) {
			defer wg.Done()
			file, err := ioutil.TempFile("/tmp", "lingualeo")
			if err != nil {
				out <- resultFile{Error: err, Index: idx}
			}
			fd, err := os.Create(file.Name())
			if err != nil {
				out <- resultFile{Error: err, Index: idx}
			}
			defer fd.Close()
			resp, err := http.Get(url)
			if err != nil {
				out <- resultFile{Error: err, Index: idx}
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				out <- resultFile{Error: fmt.Errorf("bad status: %s", resp.Status), Index: idx}
			}
			_, err = io.Copy(fd, resp.Body)
			if err != nil {
				out <- resultFile{Error: err, Index: idx}
			}
			out <- resultFile{Filename: file.Name(), Index: idx}
		}(url, idx)
	}
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}
