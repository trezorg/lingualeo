package messages

import (
	"fmt"

	"github.com/wsxiaoys/terminal/color"
)

// Color type
type Color string

const (
	// RED color
	RED Color = "@{r}"
	// GREEN color
	GREEN Color = "@{g}"
	// YELLOW color
	YELLOW Color = "@{y}"
	// WHITE color
	WHITE Color = "@{w}"
)

// Message shows a message with color package
func Message(c Color, message string, params ...any) error {
	msg := fmt.Sprintf("%s%s", c, message)
	_, err := color.Printf(msg, params...)
	return err
}
