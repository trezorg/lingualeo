package translator

import "github.com/trezorg/lingualeo/internal/api"

type Option func(*Lingualeo) error

// WithClient sets the API client for the translator.
// This is the preferred option name for injecting the API client.
func WithClient(t api.Client) Option {
	return func(l *Lingualeo) error {
		l.Client = t
		return nil
	}
}

// WithAPIClient sets the API client for the translator.
//
// Deprecated: Use WithClient instead.
func WithAPIClient(t api.Client) Option {
	return func(l *Lingualeo) error {
		l.Client = t
		return nil
	}
}

func WithOutputer(o Outputer) Option {
	return func(l *Lingualeo) error {
		l.Outputer = o
		return nil
	}
}

func WithDownloader(d Downloader) Option {
	return func(l *Lingualeo) error {
		l.Downloader = d
		return nil
	}
}

func WithPronouncer(p Pronouncer) Option {
	return func(l *Lingualeo) error {
		l.Pronouncer = p
		return nil
	}
}
