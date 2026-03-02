package messages

import (
	"github.com/fatih/color"
)

// Color type
type Color int

const (
	// RED color
	RED Color = iota
	// GREEN color
	GREEN
	// YELLOW color
	YELLOW
	// WHITE
	WHITE
)

// Pre-initialized color instances to avoid creating them on each call
var colors = map[Color]*color.Color{
	RED:    color.New(color.FgRed),
	GREEN:  color.New(color.FgGreen),
	YELLOW: color.New(color.FgYellow),
	WHITE:  color.New(color.FgWhite),
}

// Message shows a message with color package
func Message(c Color, message string, params ...any) error {
	col, ok := colors[c]
	if !ok {
		col = color.New(color.Reset)
	}
	_, err := col.Printf(message, params...)
	return err
}
