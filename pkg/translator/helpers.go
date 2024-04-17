package translator

import (
	"os/exec"
	"strings"
	"unicode"
)

func PlaySound(player string, url string) error {
	parts := strings.Split(player, " ")
	playerExec := parts[0]
	params := append(parts[1:], url)
	cmd := exec.Command(playerExec, params...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func isCommandAvailable(name string) bool {
	execName := strings.Split(name, " ")[0]
	_, err := exec.LookPath(execName)
	return err == nil
}

func isRussianWord(s string) bool {
	for _, symbol := range s {
		if !unicode.Is(unicode.Cyrillic, symbol) && !unicode.Is(unicode.Number, symbol) {
			return false
		}
	}
	return true
}

func isEnglishWord(s string) bool {
	for _, symbol := range s {
		if !unicode.Is(unicode.Latin, symbol) && !unicode.Is(unicode.Number, symbol) {
			return false
		}
	}
	return true
}
