package player

import (
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

func (p Player) Play(url string) error {
	params := append(p.params[:len(p.params):len(p.params)], url)
	cmd := exec.Command(p.player, params...)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
