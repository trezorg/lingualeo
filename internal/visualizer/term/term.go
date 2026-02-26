package term

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/BourgeoisBear/rasterm"
)

type GraphicMode string

const (
	Sixel        GraphicMode = "sixel"
	Iterm        GraphicMode = "iterm"
	Kitty        GraphicMode = "kitty"
	Unknown      GraphicMode = "unknown"
	imageTimeout             = 30 * time.Second
)

var (
	errBadStatus       = errors.New("bad status")
	errCannotReadURL   = errors.New("cannot read URL")
	errCannotReadImage = errors.New("cannot read image from url")
	errNotPaletted     = errors.New("not paletted image, skipping")
)

func Mode() GraphicMode {
	if rasterm.IsKittyCapable() {
		return Kitty
	}
	var err error
	var ok bool
	if ok, err = rasterm.IsSixelCapable(); ok {
		return Sixel
	}
	if err != nil {
		slog.Error("error checking sixel capable term", "error", err)
	}

	if rasterm.IsItermCapable() {
		return Iterm
	}
	return Unknown
}

func showImage(w io.Writer, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(data)
	ln := int64(len(data))

	img, format, err := image.Decode(reader)
	if err != nil {
		return err
	}

	_, err = reader.Seek(0, 0)
	if err != nil {
		return err
	}

	switch Mode() {
	case Iterm:
		// WEZ/ITERM SUPPORT ALL FORMATS, SO NO NEED TO RE-ENCODE TO PNG
		err = rasterm.ItermCopyFileInline(w, reader, ln)
	case Sixel:
		if iPaletted, bOK := img.(*image.Paletted); bOK {
			err = rasterm.SixelWriteImage(w, iPaletted)
		} else {
			err = errNotPaletted
		}
	case Kitty:
		if format == "png" {
			if err = rasterm.KittyCopyPNGInline(w, reader, rasterm.KittyImgOpts{}); err != nil {
				return err
			}
		} else {
			if err = rasterm.KittyWriteImage(w, img, rasterm.KittyImgOpts{}); err != nil {
				return err
			}
		}
	case Unknown:
		return nil
	default:
		slog.Error("unsupported graphic mode", "mode", Mode())
		return nil
	}
	if err == nil {
		if _, printErr := fmt.Fprintln(os.Stdout, ""); printErr != nil {
			slog.Error("cannot print newline", "error", printErr)
		}
	}
	return err
}

func open(ctx context.Context, u *url.URL) error {
	// Use context with timeout to prevent hanging on unresponsive servers
	ctx, cancel := context.WithTimeout(ctx, imageTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("%w: %s, %w", errCannotReadURL, u.String(), err)
	}
	resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is parsed and validated before visualizing
	if err != nil {
		return fmt.Errorf("%w: %s, %w", errCannotReadURL, u.String(), err)
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			slog.Error("cannot close response body", "error", cErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", errBadStatus, resp.Status)
	}
	if err = showImage(os.Stdout, resp.Body); err != nil {
		return fmt.Errorf("%w: %s, %w", errCannotReadImage, u.String(), err)
	}
	return nil
}

type Visualizer func(ctx context.Context, u *url.URL) error

func (v Visualizer) Show(ctx context.Context, u *url.URL) error {
	return v(ctx, u)
}

func New() Visualizer {
	return Visualizer(open)
}
