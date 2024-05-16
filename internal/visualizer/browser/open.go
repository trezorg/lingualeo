package browser

import (
	"net/url"
	"os/exec"
	"runtime"
)

func open(u *url.URL) error {
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
	return exec.Command(cmd, args...).Start()
}

type Visualizer func(u *url.URL) error

func (v Visualizer) Show(u *url.URL) error {
	return v(u)
}

func New() Visualizer {
	return Visualizer(open)
}
