package translator

import (
	"fmt"
	"net/url"
)

// Visualizer interface
//
//go:generate mockery
type Visualizer interface {
	Show(u *url.URL) error
}
type VisualiseType string

const (
	Default VisualiseType = "default"
	Term    VisualiseType = "term"
)

var (
	VisualiseTypes       = []VisualiseType{Default, Term}
	VisualiseTypeDefault = Default
)

func (v *VisualiseType) Set(value string) error {
	vt := VisualiseType(value)
	switch vt {
	case Default, Term:
		*v = vt
		return nil
	default:
		return fmt.Errorf("allowed: %s", VisualiseTypes)
	}
}

func (v *VisualiseType) String() string {
	return string(*v)
}
