package translator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trezorg/lingualeo/internal/api"
)

func TestBootstrapInjectsRuntimeDependencies(t *testing.T) {
	t.Parallel()

	app := Lingualeo{Config: Config{Email: "user@example.com", Password: "secret"}}

	err := Bootstrap(&app)
	require.NoError(t, err)
	require.NoError(t, app.Validate())
	require.NotNil(t, app.Client)
	require.NotNil(t, app.Downloader)
	require.NotNil(t, app.Outputer)
	require.Nil(t, app.Pronouncer)
}

func TestBootstrapInjectsPronouncerWhenSoundEnabled(t *testing.T) {
	t.Parallel()

	app := Lingualeo{Config: Config{Email: "user@example.com", Password: "secret", Sound: true, Player: "echo"}}

	err := Bootstrap(&app)
	require.NoError(t, err)
	require.NotNil(t, app.Pronouncer)
	require.NoError(t, app.Validate())
}

func TestBootstrapReturnsOutputerError(t *testing.T) {
	t.Parallel()

	app := Lingualeo{Config: Config{Email: "user@example.com", Password: "secret", Visualise: true, VisualiseType: VisualiseType("broken")}}

	err := Bootstrap(&app)
	require.Error(t, err)
	require.ErrorContains(t, err, "create outputer")
}

func TestValidateRequiresRuntimeDependencies(t *testing.T) {
	t.Parallel()

	t.Run("missing client and outputer", func(t *testing.T) {
		app := Lingualeo{}

		err := app.Validate()
		require.Error(t, err)
		require.True(t, errors.Is(err, errMissingClient))
		require.True(t, errors.Is(err, errMissingOutputer))
	})

	t.Run("missing pronouncer when sound enabled", func(t *testing.T) {
		app := Lingualeo{Config: Config{Sound: true}, Outputer: Output{}}

		err := app.Validate()
		require.Error(t, err)
		require.True(t, errors.Is(err, errMissingClient))
		require.True(t, errors.Is(err, errMissingPronouncer))
	})

	t.Run("missing downloader when sound download enabled", func(t *testing.T) {
		app := Lingualeo{
			Client:     apiMockClient{},
			Pronouncer: noopPronouncer{},
			Outputer:   Output{},
			Config: Config{
				Sound:             true,
				DownloadSoundFile: true,
			},
		}

		err := app.Validate()
		require.Error(t, err)
		require.True(t, errors.Is(err, errMissingDownloader))
	})

	t.Run("valid runtime dependencies", func(t *testing.T) {
		app := Lingualeo{
			Client:     apiMockClient{},
			Downloader: noopDownloader{},
			Pronouncer: noopPronouncer{},
			Outputer:   Output{},
			Config: Config{
				Sound:             true,
				DownloadSoundFile: true,
			},
		}

		require.NoError(t, app.Validate())
	})
}

type apiMockClient struct{}

func (apiMockClient) TranslateWord(_ context.Context, _ string) api.OperationResult {
	return api.OperationResult{}
}

func (apiMockClient) AddWord(_ context.Context, _, _ string) api.OperationResult {
	return api.OperationResult{}
}

func (apiMockClient) Auth(_ context.Context) error {
	return nil
}

type noopDownloader struct{}

func (noopDownloader) Download(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (noopDownloader) Remove(_ string) error {
	return nil
}

type noopPronouncer struct{}

func (noopPronouncer) Play(_ context.Context, _ string) error {
	return nil
}
