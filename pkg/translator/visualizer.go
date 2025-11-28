package translator

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Visualizer interface
//
//go:generate mockery
type Visualizer interface {
	Show(ctx context.Context, u *url.URL) error
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

var errVisualiseTypeAllowed = errors.New("allowed visualise types")

func (v *VisualiseType) Set(value string) error {
	vt := VisualiseType(value)
	switch vt {
	case Default, Term:
		*v = vt
		return nil
	default:
		return fmt.Errorf("%w: %v", errVisualiseTypeAllowed, VisualiseTypes)
	}
}

func (v *VisualiseType) String() string {
	return string(*v)
}
