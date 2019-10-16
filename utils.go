package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"unicode"

	"github.com/wsxiaoys/terminal/color"
)

// Value blank interface
type Value interface{}

func getJSON(body io.ReadCloser, target interface{}) error {
	defer func() {
		err := body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	return json.NewDecoder(body).Decode(target)
}

func getJSONFromString(body *string, target interface{}) error {
	return json.Unmarshal([]byte(*body), &target)
}

func readBody(resp *http.Response) (*string, error) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := string(data)
	return &res, nil
}

func unique(strSlice []string) []string {
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

func failIfError(err error) {
	if err != nil {
		printColorString("r", fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
}

func playSound(player string, url string) error {
	cmd := exec.Command(player, url)
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
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func printColorString(clr string, text string) {
	str := fmt.Sprintf("@{%s}%s\n", clr, text)
	_, err := color.Printf(str)
	if err != nil {
		log.Error(err)
	}
}

func printColorStringF(clr string, fmtText string, params ...interface{}) {
	text := fmt.Sprintf(fmtText, params...)
	str := fmt.Sprintf("@{%s}%s\n", clr, text)
	_, err := color.Printf(str)
	if err != nil {
		log.Error(err)
	}
}

func printTranslate(result lingualeoResult) {
	var strTitle string
	if result.Exists {
		strTitle = "existing"
	} else {
		strTitle = "new"
	}
	_, err := color.Printf("@{r}Found %s word:\n", strTitle)
	if err != nil {
		log.Error(err)
	}
	_, err = color.Printf("@{g}['%s'] (%s)\n", result.Word, result.Transcription)
	if err != nil {
		log.Error(err)
	}
	for _, word := range result.Words {
		printColorString("y", word)
	}
}

func printAddedTranslation(result lingualeoResult) {
	var strTitle string
	if result.Exists {
		strTitle = "Updated existing"
	} else {
		strTitle = "Added new"
	}
	_, err := color.Printf("@{r}%s word: ", strTitle)
	if err != nil {
		log.Error(err)
	}
	_, err = color.Printf("@{g}['%s']\n", result.Word)
	if err != nil {
		log.Error(err)
	}
}

func fileExists(name string) bool {
	stat, err := os.Stat(name)
	return !os.IsNotExist(err) && !stat.IsDir()
}

func getUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func insertIntoSlice(slice []Value, pos int, value Value) []Value {
	s := append(slice, new(Value))
	copy(s[pos+1:], s[pos:])
	s[pos] = value
	return s
}

func capitalize(s string) string {
	for index, value := range s {
		return string(unicode.ToUpper(value)) + s[index+1:]
	}
	return ""
}
