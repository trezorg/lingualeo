package player

import (
	"context"
	"os/exec"
	"syscall"
	"time"

	"github.com/google/shlex"
)

const (
	defaultShutdownTimeout = 2 * time.Second
)

type Player struct {
	player          string
	params          []string
	shutdownTimeout time.Duration
}

// Option is a functional option for Player configuration.
type Option func(*Player)

// WithShutdownTimeout sets the timeout for graceful shutdown.
func WithShutdownTimeout(d time.Duration) Option {
	return func(p *Player) {
		p.shutdownTimeout = d
	}
}

func ParseCommand(command string) (string, []string) {
	parts, err := shlex.Split(command)
	if err != nil || len(parts) == 0 {
		return "", []string{}
	}

	return parts[0], parts[1:]
}

func New(player string, opts ...Option) Player {
	execName, params := ParseCommand(player)
	p := Player{
		player:          execName,
		params:          params,
		shutdownTimeout: defaultShutdownTimeout,
	}
	for _, opt := range opts {
		opt(&p)
	}
	return p
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
		case <-time.After(p.shutdownTimeout):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			return <-done
		}
	}
}
