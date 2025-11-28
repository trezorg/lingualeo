package browser

import (
	"context"
	"net/url"
	"os/exec"
	"runtime"
)

func open(ctx context.Context, u *url.URL) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, u.String())
	return exec.CommandContext(ctx, cmd, args...).Start()
}

type Visualizer func(ctx context.Context, u *url.URL) error

func (v Visualizer) Show(ctx context.Context, u *url.URL) error {
	return v(ctx, u)
}

func New() Visualizer {
	return Visualizer(open)
}
