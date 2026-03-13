package translator

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/internal/validator"
)

var (
	errNoWords                 = errors.New("there are no words to translate")
	errAddCustomTranslation    = errors.New("custom translation requires exactly one word")
	errConfigFileMissing       = errors.New("config file is missing or invalid")
	errEmailArgumentMissing    = errors.New("email argument is missing")
	errEmailInvalid            = errors.New("email argument is invalid")
	errPasswordArgumentMissing = errors.New("password argument is missing")
	errPasswordPromptNonTTY    = errors.New("cannot prompt for password from non-terminal stdin")

	// ErrHelpOrVersionShown is returned when --help or --version flag is passed.
	// The caller should treat this as a successful exit (os.Exit(0)).
	ErrHelpOrVersionShown = errors.New("help or version shown")
)

func (l *Lingualeo) checkConfig() error {
	if len(l.ConfigPath) > 0 {
		filename, _ := filepath.Abs(l.ConfigPath)
		if !files.Exists(filename) {
			return fmt.Errorf("%w: %s", errConfigFileMissing, filename)
		}
	}
	return nil
}

func (l *Lingualeo) checkArgs() error {
	if len(l.Email) == 0 {
		return errEmailArgumentMissing
	}
	if err := validator.ValidateEmail(l.Email); err != nil {
		return fmt.Errorf("%w: %w", errEmailInvalid, err)
	}
	if len(l.Password) == 0 {
		return errPasswordArgumentMissing
	}
	if len(l.Words) == 0 {
		return errNoWords
	}
	return nil
}
