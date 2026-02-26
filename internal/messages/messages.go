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

// Message shows a message with color package
func Message(c Color, message string, params ...any) error {
	var col *color.Color
	switch c {
	case RED:
		col = color.New(color.FgRed)
	case GREEN:
		col = color.New(color.FgGreen)
	case YELLOW:
		col = color.New(color.FgYellow)
	case WHITE:
		col = color.New(color.FgWhite)
	default:
		col = color.New(color.Reset)
	}
	_, err := col.Printf(message, params...)
	return err
}
