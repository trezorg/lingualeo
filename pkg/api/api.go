package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/trezorg/lingualeo/internal/logger"
	"github.com/trezorg/lingualeo/pkg/channel"

	"golang.org/x/net/publicsuffix"
)

// API structure represents API request
type API struct {
	client   *http.Client
	Email    string
	Password string
	Debug    bool
}

func checkAuthError(body []byte) error {
	if len(body) == 0 {
		return nil
	}
	res := apiError{}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.ErrorCode != 0 {
		return fmt.Errorf("%s: Status code: %d", res.ErrorMsg, res.ErrorCode)
	}
	return nil
}

// New constructor
func New(email string, password string, debug bool) (*API, error) {
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
		"email":    api.Email,
		"password": api.Password,
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return nil
	}
	responseBody, err := request("POST", authURL, api.client, jsonValue, "", api.Debug)
	if err != nil {
		return err
	}
	return checkAuthError(responseBody)
}

func debugRequest(request *http.Request) {
	dump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		slog.Error("cannot dump http request", "error", err)
	} else {
		logger.SetHandler(logger.DebugHandler())
		slog.Debug(string(dump))
		logger.SetHandler(logger.DefaultHandler())
	}
}

func debugResponse(response *http.Response) {
	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		slog.Error("cannot dump http response", "error", err)
	} else {
		logger.SetHandler(logger.DebugHandler())
		slog.Debug(string(dump))
		logger.SetHandler(logger.DefaultHandler())
	}
}

func request(method string, url string, client *http.Client, body []byte, query string, debug bool) ([]byte, error) {
	var requestBody io.Reader
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

	if debug {
		debugRequest(req)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if debug {
		debugResponse(resp)
	}
	defer func() {
		dErr := resp.Body.Close()
		if dErr != nil {
			slog.Error("cannot close response body", "error", dErr)
		}
	}()
	responseBody, err := readBody(resp)
	if err != nil {
		slog.Error("cannot read response body", "error", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"response status code: %d\nbody:\n%s",
			resp.StatusCode,
			string(responseBody),
		)
	}
	return responseBody, err
}

func (api *API) translateRequest(word string) ([]byte, error) {
	values := map[string]interface{}{
		"text":       word,
		"apiVersion": apiVersion,
		"ctx": map[string]interface{}{
			"config": map[string]interface{}{
				"isCheckData": true,
				"isLogging":   true,
			},
		},
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	return request("POST", translateURL, api.client, jsonValue, "", api.Debug)
}

func (api *API) addRequest(word string, translate []string) ([]byte, error) {
	values := map[string]string{
		"word":  word,
		"tword": strings.Join(translate, ", "),
		"port":  "1001",
	}
	jsonValue, _ := json.Marshal(values)
	return request("POST", addWordURL, api.client, jsonValue, "", api.Debug)
}

func (api *API) TranslateWord(word string) OperationResult {
	body, err := api.translateRequest(word)
	if err != nil {
		return OperationResult{Error: err}
	}
	return opResultFromBody(word, body)
}

func (api *API) AddWord(word string, translate []string) OperationResult {
	body, err := api.addRequest(word, translate)
	if err != nil {
		return OperationResult{Error: err}
	}
	return opResultFromBody(word, body)
}

// TranslateWords transate words from string channel
func (api *API) TranslateWords(ctx context.Context, results <-chan string) <-chan OperationResult {
	out := make(chan OperationResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for word := range channel.OrDone(ctx, results) {
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

// AddWords add words
func (api *API) AddWords(ctx context.Context, results <-chan Result) <-chan OperationResult {
	out := make(chan OperationResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for res := range channel.OrDone(ctx, results) {
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
