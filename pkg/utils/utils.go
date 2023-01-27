package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"unicode"

	"github.com/trezorg/lingualeo/pkg/logger"
	"github.com/trezorg/lingualeo/pkg/messages"
)

// Value blank interface
type Value interface{}

func GetJSON(body io.ReadCloser, target interface{}) error {
	defer func() {
		err := body.Close()
		if err != nil {
			logger.Error(err)
		}
	}()
	return json.NewDecoder(body).Decode(target)
}

func GetJSONFromString(body *string, target interface{}) error {
	return json.Unmarshal([]byte(*body), &target)
}

func ReadBody(resp *http.Response) (*string, error) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Error(err)
		}
	}()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := string(data)
	return &res, nil
}

func Unique(strSlice []string) []string {
	keys := make(map[string]bool, len(strSlice))
	var list []string
	for _, entry := range strSlice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func FailIfError(err error) {
	if err != nil {
		_ = messages.Message(messages.RED, fmt.Sprintf("Error: %v\n", err))
		os.Exit(1)
	}
}

func PlaySound(ctx context.Context, player string, url string) error {
	parts := strings.Split(player, " ")
	playerExec := parts[0]
	params := append(parts[1:], url)
	cmd := exec.CommandContext(ctx, playerExec, params...)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func IsCommandAvailable(name string) bool {
	execName := strings.Split(name, " ")[0]
	_, err := exec.LookPath(execName)
	return err == nil
}

func FileExists(name string) bool {
	stat, err := os.Stat(name)
	return !os.IsNotExist(err) && !stat.IsDir()
}

func GetUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func InsertIntoSlice(slice []Value, pos int, value Value) []Value {
	s := append(slice, new(Value))
	copy(s[pos+1:], s[pos:])
	s[pos] = value
	return s
}

func Capitalize(s string) string {
	for index, value := range s {
		return string(unicode.ToUpper(value)) + s[index+1:]
	}
	return ""
}

func IsRussianWord(s string) bool {
	for _, symbol := range s {
		if !(unicode.Is(unicode.Cyrillic, symbol) || unicode.Is(unicode.Number, symbol)) {
			return false
		}
	}
	return true
}

func IsEnglishWord(s string) bool {
	for _, symbol := range s {
		if !(unicode.Is(unicode.Latin, symbol) || unicode.Is(unicode.Number, symbol)) {
			return false
		}
	}
	return true
}
