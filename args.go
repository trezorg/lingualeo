package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/alyu/configparser"
	"gopkg.in/yaml.v2"
)

type arrayFlags []string

func (s *arrayFlags) String() string {
	return strings.Join(*s, ", ")
}

func (s *arrayFlags) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func prepareArgs() lingualeoArgs {
	var translateFlag arrayFlags
	emailPtr := flag.String("e", "", "Lingualeo email")
	passwordPtr := flag.String("p", "", "Lingualeo password")
	configPtr := flag.String("c", "", `
Config file. Either in plain ini, yaml or json format.

Plain format example:

email = email@gmail.com
password = password
add = false
sound = true
player = mplayer

Yaml format example:

email: email@gmail.com
password: password
add: false
sound: true
player: mplayer

JSON format example:

{
    "email": "email@gmail.com",
    "password": "password",
    "add": false,
    "sound": true,
    "player": "mplayer"
}

Default config files are: ~/lingualeo.[conf|yml|yaml|json]`)
	playerPtr := flag.String("m", "", "Media player for word pronunciation")
	forcePtr := flag.Bool("f", false, "Force add to lingualeo dictionary")
	addPtr := flag.Bool("a", false, "Add to lingualeo dictionary")
	soundPtr := flag.Bool("s", false, "Play words pronunciation")
	logPrettyPrintPtr := flag.Bool("pr", false, "Log pretty print")
	logLevel := flag.String("l", "INFO", "Log level")
	flag.Var(&translateFlag, "t", "Custom translation")
	flag.Parse()
	words := flag.Args()
	return lingualeoArgs{
		*emailPtr,
		*passwordPtr,
		*configPtr,
		*playerPtr,
		words,
		translateFlag,
		*forcePtr,
		*addPtr,
		*soundPtr,
		*logLevel,
		*logPrettyPrintPtr,
	}
}

func setStringOption(args *lingualeoArgs, name string, options map[string]string) {
	value, exists := options[strings.ToLower(name)]
	if exists && len(value) > 0 {
		reflect.ValueOf(args).Elem().FieldByName(name).SetString(value)
	}
}

func setBoolOption(args *lingualeoArgs, name string, options map[string]string) error {
	value, exists := options[strings.ToLower(name)]
	if exists {
		res, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		reflect.ValueOf(args).Elem().FieldByName(name).SetBool(res)
	}
	return nil
}

func readIniConfig(args *lingualeoArgs, filename string) error {
	config, err := configparser.Read(filename)
	if err != nil {
		return err
	}
	sections, err := config.AllSections()
	if err != nil {
		return err
	}
	options := sections[0].Options()
	for _, attr := range []string{"Email", "Password", "Player", "LogLevel"} {
		setStringOption(args, attr, options)
	}
	for _, attr := range []string{"Force", "Add", "Sound", "LogPrettyPrint"} {
		err := setBoolOption(args, attr, options)
		if err != nil {
			return err
		}
	}
	args.Config = filename
	return nil
}

func readYamlConfig(args *lingualeoArgs, filename string) error {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, args)
	if err != nil {
		return err
	}
	return nil
}

func readJSONConfig(args *lingualeoArgs, filename string) error {
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonFile, args)
	if err != nil {
		return err
	}
	return nil
}

func readConfig(args *lingualeoArgs, filename string) error {
	extension := filepath.Ext(filename)
	if extension == ".yml" || extension == ".yaml" {
		return readYamlConfig(args, filename)
	}
	if extension == ".json" {
		return readJSONConfig(args, filename)
	}
	return readIniConfig(args, filename)
}

func readConfigs(filename *string) (*lingualeoArgs, error) {
	home, err := getUserHome()
	if err != nil {
		return nil, err
	}
	var configs []string
	var homeConfigFile string
	var currentConfigFile string
	for _, configFilename := range defaultConfigFiles {
		homeConfigFile, _ = filepath.Abs(filepath.Join(home, configFilename))
		currentConfigFile, _ = filepath.Abs(configFilename)
		for _, fullConfigFileName := range [2]string{homeConfigFile, currentConfigFile} {
			if fileExists(fullConfigFileName) {
				configs = append(configs, fullConfigFileName)
			}
		}
	}
	if len(*filename) > 0 {
		argsConfig, _ := filepath.Abs(*filename)
		configs = append(configs, argsConfig)
	}
	configs = unique(configs)
	args := &lingualeoArgs{}
	for _, name := range configs {
		err = readConfig(args, name)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}

func checkConfig(args *lingualeoArgs) error {
	if len(args.Config) > 0 {
		filename, _ := filepath.Abs(args.Config)
		if !fileExists(filename) {
			return fmt.Errorf("there is no the config file or file is a directory: %s", filename)
		}
	}
	return nil
}

func checkArgs(args *lingualeoArgs) error {
	if len(args.Email) == 0 {
		return fmt.Errorf("mo email argument has been supplied")
	}
	if len(args.Password) == 0 {
		return fmt.Errorf("no password argument has been supplied")
	}
	if len(args.Words) == 0 {
		return fmt.Errorf("no words to translate have been supplied")
	}
	return nil
}

func mergeConfigs(args *lingualeoArgs, configArgs *lingualeoArgs) *lingualeoArgs {
	if len(args.Email) == 0 && len(configArgs.Email) > 0 {
		args.Email = configArgs.Email
	}
	if len(args.Password) == 0 && len(configArgs.Password) > 0 {
		args.Password = configArgs.Password
	}
	if len(args.Player) == 0 && len(configArgs.Player) > 0 {
		args.Player = configArgs.Player
	}
	if configArgs.Force {
		args.Force = configArgs.Force
	}
	if configArgs.Add {
		args.Add = configArgs.Add
	}
	if configArgs.Sound {
		args.Sound = configArgs.Sound
	}
	if configArgs.LogPrettyPrint {
		args.LogPrettyPrint = configArgs.LogPrettyPrint
	}
	if len(args.LogLevel) == 0 && len(configArgs.LogLevel) > 0 {
		args.LogLevel = configArgs.LogLevel
	}
	return args
}
