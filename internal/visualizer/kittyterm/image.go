package kittyterm

import (
	"fmt"
	"image"
	_ "image/gif" // gif
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/dolmen-go/kittyimg"
)

func showFromReader(r io.Reader) error {
	img, _, err := image.Decode(r)
	if err != nil {
		return fmt.Errorf("cannot read image from reader. %w", err)
	}
	return kittyimg.Fprintln(os.Stdout, img)
}

func open(u *url.URL) error {
	resp, err := http.Get(u.String())
	if err != nil {
		return fmt.Errorf("cannot read URL: %s, %w", u, err)
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			slog.Error("cannot close response body", "error", cErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	if err = showFromReader(resp.Body); err != nil {
		return fmt.Errorf("cannot read image from url: %s. %w", u, err)
	}
	return nil
}

type Visualizer func(u *url.URL) error

func (v Visualizer) Show(u *url.URL) error {
	return v(u)
}

func New() Visualizer {
	return Visualizer(open)
}
