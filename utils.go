package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/wsxiaoys/terminal/color"
)

func getJSON(body io.ReadCloser, target interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(target)
}

func getJSONFromString(body string, target interface{}) error {
	return json.Unmarshal([]byte(body), &target)
}

func readBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unique(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
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
	color.Printf(str)
}

func printTranslate(result *lingualeoResult) {
	var strTitle string
	if result.Exists {
		strTitle = "existing"
	} else {
		strTitle = "new"
	}
	color.Printf("@{r}Found %s word:\n", strTitle)
	color.Printf("@{g}['%s'] (%s)\n", result.Word, result.Transcription)
	for _, word := range result.Words {
		printColorString("y", word)
	}
}

func printAddTranslate(result *lingualeoResult) {
	var strTitle string
	if result.Exists {
		strTitle = "Updated existing"
	} else {
		strTitle = "Added new"
	}
	color.Printf("@{r}%s word: ", strTitle)
	color.Printf("@{g}['%s']\n", result.Word)
}
