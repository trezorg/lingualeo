package translator

import (
	"errors"
	"fmt"

	"github.com/trezorg/lingualeo/internal/api"
	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/internal/httpclient"
	"github.com/trezorg/lingualeo/internal/player"
)

var (
	errMissingClient     = errors.New("api client is required")
	errMissingOutputer   = errors.New("outputer is required")
	errMissingPronouncer = errors.New("pronouncer is required when sound is enabled")
	errMissingDownloader = errors.New("downloader is required when downloading sound")
)

func Bootstrap(app *Lingualeo) error {
	httpClient, err := httpclient.NewWithJar(
		httpclient.Config{
			MaxIdleConns:        app.MaxIdleConns,
			MaxIdleConnsPerHost: app.MaxIdleConnsPerHost,
		},
		app.MaxRedirects,
	)
	if err != nil {
		return fmt.Errorf("create HTTP client: %w", err)
	}

	app.Client = api.New(app.Email, app.Password, app.Debug, app.APIClientConfig(), httpClient)
	app.Downloader = files.New(httpClient)

	if app.Sound {
		app.Pronouncer = player.New(app.Player, player.WithShutdownTimeout(app.PlayerShutdownTimeout))
	}

	outputer, err := NewOutputer(app.Visualise, app.VisualiseType)
	if err != nil {
		return fmt.Errorf("create outputer: %w", err)
	}
	app.Outputer = outputer

	return nil
}
