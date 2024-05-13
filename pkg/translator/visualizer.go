package translator

import "net/url"

// Visualizer interface
//
//go:generate mockery
type Visualizer interface {
	Show(u *url.URL) error
}
