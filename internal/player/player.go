package player

import (
	"context"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	separator       = " "
	shutdownTimeout = 2 * time.Second
)

type Player struct {
	player string
	params []string
}

func New(player string) Player {
	parts := strings.Split(player, separator)
	playerExec := parts[0]
	params := parts[1:]
	return Player{
		params: params,
		player: playerExec,
	}
}

func (p Player) Play(ctx context.Context, url string) error {
	params := append([]string{}, p.params...)
	params = append(params, url)
	//nolint:gosec // the player command is invoked deliberately with user-provided URLs
	cmd := exec.CommandContext(ctx, p.player, params...)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for process in goroutine, handle context cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Graceful shutdown: SIGTERM first, then SIGKILL after timeout
		if cmd.Process != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
		select {
		case <-done:
			return ctx.Err()
		case <-time.After(shutdownTimeout):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			return <-done
		}
	}
}
