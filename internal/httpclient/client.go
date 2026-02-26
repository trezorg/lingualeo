package httpclient

import (
	"context"
	"net/http"
	"time"
)

const (
	DefaultTimeout          = 30 * time.Second
	DefaultMaxIdleConns     = 10
	DefaultMaxIdleConnsHost = 10
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

// WithTimeout wraps a request with context timeout.
// Returns the new request and cancel function that must be called.
func WithTimeout(ctx context.Context, req *http.Request, timeout time.Duration) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	return req.WithContext(ctx), cancel
}
