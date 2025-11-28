package player

import (
	"context"
	"os/exec"
	"strings"
)

const (
	separator = " "
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
	return cmd.Wait()
}
