package translator

import (
	"context"
	"os/exec"

	"github.com/trezorg/lingualeo/internal/player"
)

// Pronouncer interface
//
//go:generate mockery
type Pronouncer interface {
	Play(ctx context.Context, url string) error
}

func isCommandAvailable(name string) bool {
	execName, _ := player.ParseCommand(name)
	if execName == "" {
		return false
	}
	_, err := exec.LookPath(execName)
	return err == nil
}
