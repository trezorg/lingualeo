package api

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/trezorg/lingualeo/internal/httpclient"

	"github.com/avast/retry-go/v5"
	"golang.org/x/net/publicsuffix"
)

// Default configuration values
const (
	defaultMaxRedirects = 10
	defaultMaxAttempts  = 3
	defaultInitialWait  = 500 * time.Millisecond
	defaultMaxWait      = 5 * time.Second
	addWordPort         = "1001"
)

var (
	errAPIAuth           = errors.New("api authentication error")
	errAPIResponseStatus = errors.New("unexpected response status code")
	errAPIRedirectLimit  = errors.New("too many redirects")
	errAPIRequestTimeout = errors.New("api request timeout")
)

// RetryConfig holds retry configuration for API requests.
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// Config holds configuration for API client.
type Config struct {
	Timeout             time.Duration
	MaxRedirects        int
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	Retry               RetryConfig
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Timeout:             httpclient.DefaultTimeout,
		MaxRedirects:        defaultMaxRedirects,
		MaxIdleConns:        httpclient.DefaultMaxIdleConns,
		MaxIdleConnsPerHost: httpclient.DefaultMaxIdleConnsHost,
		Retry: RetryConfig{
			MaxAttempts: defaultMaxAttempts,
			InitialWait: defaultInitialWait,
			MaxWait:     defaultMaxWait,
		},
	}
}

// API structure represents API request
// Client interface defines the contract for Lingualeo API operations.
// Use this interface for mocking API calls in tests.
//
//go:generate mockery
type Client interface {
	TranslateWord(ctx context.Context, word string) OperationResult
	AddWord(ctx context.Context, word string, translate string) OperationResult
}

type API struct {
	client      *http.Client
	Email       string
	Password    string //nolint:gosec // false positive: credential field name is intentional
	Debug       bool
	timeout     time.Duration
	retryConfig RetryConfig
}

type requestParams struct {
	method string
	url    string
	body   []byte
	query  string
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
func New(ctx context.Context, email string, password string, debug bool, cfg Config) (*API, error) {
	// Apply defaults for zero values
	cfg.Timeout = cmp.Or(cfg.Timeout, httpclient.DefaultTimeout)
	cfg.MaxRedirects = cmp.Or(cfg.MaxRedirects, defaultMaxRedirects)
	cfg.MaxIdleConns = cmp.Or(cfg.MaxIdleConns, httpclient.DefaultMaxIdleConns)
	cfg.MaxIdleConnsPerHost = cmp.Or(cfg.MaxIdleConnsPerHost, httpclient.DefaultMaxIdleConnsHost)
	cfg.Retry.MaxAttempts = cmp.Or(cfg.Retry.MaxAttempts, defaultMaxAttempts)
	cfg.Retry.InitialWait = cmp.Or(cfg.Retry.InitialWait, defaultInitialWait)
	cfg.Retry.MaxWait = cmp.Or(cfg.Retry.MaxWait, defaultMaxWait)

	client, err := prepareClient(cfg)
	if err != nil {
		return nil, err
	}
	api := &API{
		Email:       email,
		Password:    password,
		Debug:       debug,
		client:      client,
		timeout:     cfg.Timeout,
		retryConfig: cfg.Retry,
	}
	return api, api.auth(ctx)
}

func prepareClient(cfg Config) (*http.Client, error) {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		return nil, err
	}
	netTransport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
	}

	client := &http.Client{
		Jar:       jar,
		Transport: netTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.MaxRedirects {
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

func (a *API) auth(ctx context.Context) error {
	values := map[string]string{
		"email":    a.Email,
		"password": a.Password,
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return err
	}
	responseBody, err := a.request(ctx, requestParams{
		method: "POST",
		url:    authURL,
		body:   jsonValue,
	})
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
		//nolint:gosec // debug output intentionally logs full HTTP request dump
		slog.Debug(string(dump))
	}
}

func debugResponse(response *http.Response) {
	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		slog.Error("cannot dump http response", "error", err)
	} else {
		//nolint:gosec // debug output intentionally logs full HTTP response dump
		slog.Debug(string(dump))
	}
}

// isRetryable checks if an error or status code should trigger a retry.
func isRetryable(err error, statusCode int) bool {
	// Network errors are retryable
	if err != nil {
		return true
	}
	// 5xx server errors are retryable
	if statusCode >= 500 && statusCode < 600 {
		return true
	}
	// 429 Too Many Requests is retryable
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	return false
}

func (a *API) request(ctx context.Context, params requestParams) ([]byte, error) {
	var body []byte
	var lastErr error

	retrier := retry.New(
		retry.Context(ctx),
		retry.Attempts(uint(a.retryConfig.MaxAttempts)), //nolint:gosec // G115: safe conversion, value is always small positive int
		retry.Delay(a.retryConfig.InitialWait),
		retry.MaxDelay(a.retryConfig.MaxWait),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			slog.Debug("retrying request", "attempt", n+1, "error", err, "url", params.url)
		}),
	)

	err := retrier.Do(
		func() error {
			var statusCode int
			body, statusCode, lastErr = a.doRequest(ctx, params)
			if lastErr != nil {
				return lastErr
			}
			if statusCode != http.StatusOK && isRetryable(nil, statusCode) {
				return fmt.Errorf("%w: status code: %d", errAPIResponseStatus, statusCode)
			}
			if statusCode != http.StatusOK {
				return retry.Unrecoverable(fmt.Errorf(
					"%w: status code: %d\nbody:\n%s",
					errAPIResponseStatus,
					statusCode,
					string(body),
				))
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}
	return body, nil
}

// doRequest performs a single HTTP request without retry logic.
func (a *API) doRequest(ctx context.Context, params requestParams) ([]byte, int, error) {
	ctx, cancel := context.WithTimeoutCause(ctx, a.timeout, errAPIRequestTimeout)
	defer cancel()

	var requestBody io.Reader
	if len(params.body) > 0 {
		requestBody = bytes.NewBuffer(params.body)
	}
	req, err := http.NewRequestWithContext(ctx, params.method, params.url, requestBody)
	if err != nil {
		return nil, 0, err
	}
	if len(params.query) > 0 {
		req.URL.RawQuery = params.query
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if len(params.body) > 0 {
		req.Header.Set("Content-Length", strconv.Itoa(len(params.body)))
	}

	for key, values := range agentHeaders {
		for _, header := range values {
			req.Header.Add(key, header)
		}
	}

	if a.Debug {
		debugRequest(req)
	}
	resp, err := a.client.Do(req) //nolint:gosec // URL is internal API constant configured by the application
	if err != nil {
		return nil, 0, err
	}
	if a.Debug {
		debugResponse(resp)
	}
	defer func() {
		dErr := resp.Body.Close()
		if dErr != nil {
			slog.Error("cannot close response body", "error", dErr)
		}
	}()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("cannot read response body", "error", err)
		return nil, resp.StatusCode, err
	}
	return responseBody, resp.StatusCode, nil
}

func (a *API) translateRequest(ctx context.Context, word string) ([]byte, error) {
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
	return a.request(ctx, requestParams{
		method: "POST",
		url:    translateURL,
		body:   jsonValue,
	})
}

func (a *API) addRequest(ctx context.Context, word string, translate string) ([]byte, error) {
	values := map[string]string{
		"word":  word,
		"tword": translate,
		"port":  addWordPort,
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return a.request(ctx, requestParams{
		method: "POST",
		url:    addWordURL,
		body:   jsonValue,
	})
}

func (a *API) TranslateWord(ctx context.Context, word string) OperationResult {
	body, err := a.translateRequest(ctx, word)
	if err != nil {
		return OperationResult{Error: err}
	}
	return opResultFromBody(word, body)
}

func (a *API) AddWord(ctx context.Context, word string, translate string) OperationResult {
	body, err := a.addRequest(ctx, word, translate)
	if err != nil {
		return OperationResult{Error: err}
	}
	return opResultFromBody(word, body)
}
