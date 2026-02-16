package translator

import "github.com/trezorg/lingualeo/internal/api"

type Option func(*Lingualeo) error

func WithTranslator(t api.Client) Option {
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
