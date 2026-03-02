package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"time"

	"golang.org/x/net/publicsuffix"
)

var errRequestTimeout = errors.New("http request timeout")

const (
	DefaultTimeout          = 30 * time.Second
	DefaultMaxIdleConns     = 10
	DefaultMaxIdleConnsHost = 10
	DefaultMaxRedirects     = 10
)

type Config struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
}

// New creates an HTTP client with connection pooling.
// Timeouts should be handled via context at call sites, not at client level.
func New(cfg Config) *http.Client {
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = DefaultMaxIdleConns
	}
	if cfg.MaxIdleConnsPerHost == 0 {
		cfg.MaxIdleConnsPerHost = DefaultMaxIdleConnsHost
	}
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		},
	}
}

// ErrRedirectLimit is returned when the redirect limit is exceeded.
var ErrRedirectLimit = errors.New("too many redirects")

// NewWithJar creates an http.Client with cookie jar, connection pooling, and redirect policy.
// This is the recommended client for API interactions that require cookie handling.
func NewWithJar(cfg Config, maxRedirects int) (*http.Client, error) {
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = DefaultMaxIdleConns
	}
	if cfg.MaxIdleConnsPerHost == 0 {
		cfg.MaxIdleConnsPerHost = DefaultMaxIdleConnsHost
	}
	if maxRedirects == 0 {
		maxRedirects = DefaultMaxRedirects
	}

	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return ErrRedirectLimit
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
	}, nil
}

// WithTimeout wraps a request with context timeout.
// Returns the new request and cancel function that must be called.
func WithTimeout(ctx context.Context, req *http.Request, timeout time.Duration) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeoutCause(ctx, timeout, errRequestTimeout)
	return req.WithContext(ctx), cancel
}
