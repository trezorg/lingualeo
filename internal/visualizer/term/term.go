package term

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/BourgeoisBear/rasterm"
)

type Mode string

const (
	Sixel   Mode = "sixel"
	Iterm   Mode = "iterm"
	Kitty   Mode = "kitty"
	Unknown Mode = "unknown"
)

func getMode() Mode {
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

	switch getMode() {
	case Iterm:

		// WEZ/ITERM SUPPORT ALL FORMATS, SO NO NEED TO RE-ENCODE TO PNG
		err = rasterm.ItermCopyFileInline(w, reader, ln)

	case Sixel:

		if iPaletted, bOK := img.(*image.Paletted); bOK {
			err = rasterm.SixelWriteImage(w, iPaletted)
		} else {
			slog.Debug("not paletted image, skipping")
		}

	case Kitty:

		if format == "png" {
			// fmt.Println("Kitty PNG Local File")
			// eF := rasterm.KittyWritePNGLocal(w, fpath, rasterm.KittyImgOpts{})
			// fmt.Println("\nKitty PNG Inline")
			if err = rasterm.KittyCopyPNGInline(w, reader, rasterm.KittyImgOpts{}); err != nil {
				return err
			}
			// err = errors.Join(eI, eF)
		} else {
			if err = rasterm.KittyWriteImage(w, img, rasterm.KittyImgOpts{}); err != nil {
				return err
			}
		}

	default:
		return nil

	}

	fmt.Println("")
	return err
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
	if err = showImage(os.Stdout, resp.Body); err != nil {
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
