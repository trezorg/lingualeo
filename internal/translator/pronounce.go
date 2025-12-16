package translator

import (
	"context"
	"os/exec"
	"strings"
)

// Pronouncer interface
//
//go:generate mockery
type Pronouncer interface {
	Play(ctx context.Context, url string) error
}

func isCommandAvailable(name string) bool {
	execName := strings.Split(name, " ")[0]
	_, err := exec.LookPath(execName)
	return err == nil
}
