package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/trezorg/lingualeo/pkg/channel"
	"github.com/trezorg/lingualeo/pkg/logger"

	"golang.org/x/net/publicsuffix"
)

type Translator interface {
	TranslateWord(word string) OpResult
	TranslateWords(ctx context.Context, results <-chan string) <-chan OpResult
	AddWord(word string, translate []string) OpResult
	AddWords(ctx context.Context, results <-chan Result) <-chan OpResult
}

type API struct {
	Email    string
	Password string
	Debug    bool
	client   *http.Client
}

func NewAPI(email string, password string, debug bool) (*API, error) {
	client, err := prepareClient()
	if err != nil {
		return nil, err
	}
	api := &API{
		Email:    email,
		Password: password,
		Debug:    debug,
		client:   client,
	}
	return api, api.auth()
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

func (api *API) auth() error {
	values := map[string]string{
		"email":              api.Email,
		"password":           api.Password,
		"type":               "login",
		"successRedirectUrl": "",
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return nil
	}
	responseBody, err := request("POST", authURL, api.client, jsonValue, "")
	if err != nil {
		return err
	}
	res := apiError{}
	err = getJSONFromString(responseBody, &res)
	if err != nil {
		return err
	}
	if res.ErrorCode != 0 {
		return fmt.Errorf("%s: Status code: %d", res.ErrorMsg, res.ErrorCode)
	}
	return nil
}

func request(method string, url string, client *http.Client, body []byte, query string) (*string, error) {
	var requestBody io.Reader = nil
	if len(body) > 0 {
		requestBody = bytes.NewBuffer(body)
	}
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		req.URL.RawQuery = query
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	}

	for key, values := range agentHeaders {
		for _, header := range values {
			req.Header.Add(key, header)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Log.Error(err)
		}
	}()
	responseBody, err := readBody(resp)
	if err != nil {
		logger.Log.Error(err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"Response status code: %d\nbody:\n%s",
			resp.StatusCode,
			*responseBody,
		)
	}
	return responseBody, err
}

func opResultFromBody(word string, body *string) OpResult {
	res := Result{Word: word}
	err := getJSONFromString(body, &res)
	if err != nil {
		res := NoResult{}
		if getJSONFromString(body, res) == nil {
			return OpResult{Error: fmt.Errorf("cannot translate word: %s", word)}
		}
		return OpResult{Error: err}
	}
	if len(res.ErrorMsg) > 0 {
		return OpResult{Error: fmt.Errorf(res.ErrorMsg)}
	}
	res.ParseTranslation()
	return OpResult{Result: res}
}

func (api *API) translateRequest(word string) (*string, error) {
	q := url.Values{}
	q.Add("word", word)
	q.Add("include_media", "1")
	q.Add("app_word_forms", "1")
	query := q.Encode()
	return request("POST", translateURL, api.client, nil, query)
}

func (api *API) addRequest(word string, translate []string) (*string, error) {
	values := map[string]string{
		"word":  word,
		"tword": strings.Join(translate, ", "),
		"port":  "1001",
	}
	jsonValue, _ := json.Marshal(values)
	return request("POST", addWordURL, api.client, jsonValue, "")
}

func (api *API) TranslateWord(word string) OpResult {
	body, err := api.translateRequest(word)
	if err != nil {
		return OpResult{Error: err}
	}
	return opResultFromBody(word, body)
}

func (api *API) AddWord(word string, translate []string) OpResult {
	body, err := api.addRequest(word, translate)
	if err != nil {
		return OpResult{Error: err}
	}
	return opResultFromBody(word, body)
}

func (api *API) TranslateWords(ctx context.Context, results <-chan string) <-chan OpResult {
	out := make(chan OpResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for word := range channel.OrStringDone(ctx, results) {
			wg.Add(1)
			go func(word string) {
				defer wg.Done()
				out <- api.TranslateWord(word)
			}(word)
		}
	}()
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func (api *API) AddWords(ctx context.Context, results <-chan Result) <-chan OpResult {
	out := make(chan OpResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for res := range OrResultDone(ctx, results) {
			wg.Add(1)
			result := res
			go func(result Result) {
				defer wg.Done()
				out <- api.AddWord(result.Word, result.Words)
			}(result)
		}
	}()
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}
