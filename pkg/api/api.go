package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/trezorg/lingualeo/internal/logger"

	"golang.org/x/net/publicsuffix"
)

const (
	maxIdleConns        = 10
	maxIdleConnsPerHost = 10
	maxRedirects        = 10
	requestTimeout      = 10 * time.Second
)

var (
	errAPIAuth           = errors.New("api authentication error")
	errAPIResponseStatus = errors.New("unexpected response status code")
	errAPIRedirectLimit  = errors.New("too many redirects")
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
		return fmt.Errorf("%w: %s (code %d)", errAPIAuth, res.ErrorMsg, res.ErrorCode)
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
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
	}

	client := &http.Client{
		Timeout:   requestTimeout,
		Jar:       jar,
		Transport: netTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errAPIRedirectLimit
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

func (a *API) auth() error {
	values := map[string]string{
		"email":    a.Email,
		"password": a.Password,
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return err
	}
	responseBody, err := request("POST", authURL, a.client, jsonValue, "", a.Debug)
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
	req, err := http.NewRequestWithContext(context.Background(), method, url, requestBody)
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"%w: status code: %d\nbody:\n%s",
			errAPIResponseStatus,
			resp.StatusCode,
			string(responseBody),
		)
	}
	return responseBody, err
}

func (a *API) translateRequest(word string) ([]byte, error) {
	values := map[string]any{
		"text":       word,
		"apiVersion": apiVersion,
		"ctx": map[string]any{
			"config": map[string]any{
				"isCheckData": true,
				"isLogging":   true,
			},
		},
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	return request("POST", translateURL, a.client, jsonValue, "", a.Debug)
}

func (a *API) addRequest(word string, translate string) ([]byte, error) {
	values := map[string]string{
		"word":  word,
		"tword": translate,
		"port":  "1001",
	}
	jsonValue, _ := json.Marshal(values)
	return request("POST", addWordURL, a.client, jsonValue, "", a.Debug)
}

func (a *API) TranslateWord(word string) OperationResult {
	body, err := a.translateRequest(word)
	if err != nil {
		return OperationResult{Error: err}
	}
	return opResultFromBody(word, body)
}

func (a *API) AddWord(word string, translate string) OperationResult {
	body, err := a.addRequest(word, translate)
	if err != nil {
		return OperationResult{Error: err}
	}
	return opResultFromBody(word, body)
}
