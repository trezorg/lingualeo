package translator

type Option func(*Lingualeo) error

func WithTranslator(t Translator) Option {
	return func(l *Lingualeo) error {
		l.Translator = t
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
