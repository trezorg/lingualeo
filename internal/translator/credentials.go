package translator

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var passwordPrompt = promptPasswordHidden

func (l *Lingualeo) promptPasswordIfNeeded() error {
	if l.Password != "" || !l.PromptPassword {
		return nil
	}

	password, err := passwordPrompt()
	if err != nil {
		return err
	}
	l.Password = password

	return nil
}

func promptPasswordHidden() (string, error) {
	fd := int(os.Stdin.Fd()) //nolint:gosec // required by x/term API
	if !term.IsTerminal(fd) {
		return "", errPasswordPromptNonTTY
	}
	_, _ = fmt.Fprint(os.Stderr, "Lingualeo password: ")
	password, err := term.ReadPassword(fd)
	_, _ = fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(password)), nil
}
