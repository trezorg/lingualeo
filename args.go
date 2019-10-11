package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"gopkg.in/yaml.v2"
)

type configType int
type decodeFunc func(data []byte, args *lingualeoArgs) error

const (
	yamlType = configType(1)
	jsonType = configType(2)
	tomlType = configType(3)
)

var (
	decodeMapping = map[configType]decodeFunc{
		yamlType: readYamlConfig,
		jsonType: readJSONConfig,
		tomlType: readTomlConfig,
	}
)

type configFileType struct {
	filename string
}

func (cf *configFileType) getType() configType {
	ext := filepath.Ext(cf.filename)
	if ext == ".yml" || ext == ".yaml" {
		return yamlType
	}
	if ext == ".json" {
		return jsonType
	}
	return tomlType
}

func (cf *configFileType) decode(data []byte, args *lingualeoArgs) error {
	return decodeMapping[cf.getType()](data, args)
}

func (cf *configFileType) decodeFile(args *lingualeoArgs) error {
	data, err := ioutil.ReadFile(cf.filename)
	if err != nil {
		return err
	}
	return cf.decode(data, args)
}

func newConfigFileType(filename string) *configFileType {
	return &configFileType{filename: filename}
}

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
Config file. Either in toml, yaml or json format.

Toml format example:

email = "email@gmail.com"
password = "password"
add = false
sound = true
player = "mplayer"

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

Default config files are: ~/lingualeo.[toml|yml|yaml|json]`)
	player := flag.String("m", "", "Media player for word pronunciation")
	force := flag.Bool("f", false, "Force add to lingualeo dictionary")
	add := flag.Bool("a", false, "Add to lingualeo dictionary")
	sound := flag.Bool("s", false, "Play words pronunciation")
	logPrettyPrint := flag.Bool("pr", false, "Log pretty print")
	translateReplaceWithAdd := flag.Bool("tr", false, "Custom translation. Replace word instead of to add")
	logLevel := flag.String("l", "INFO", "Log level")
	flag.Var(&translateFlag, "t", "Custom translation. -t word1 -t word2")
	flag.Parse()
	words := flag.Args()
	return lingualeoArgs{
		Email:                   *emailPtr,
		Password:                *passwordPtr,
		Config:                  *configPtr,
		Player:                  *player,
		Words:                   words,
		Translate:               translateFlag,
		Force:                   *force,
		Add:                     *add,
		TranslateReplaceWithAdd: *translateReplaceWithAdd,
		Sound:                   *sound,
		LogLevel:                *logLevel,
		LogPrettyPrint:          *logPrettyPrint,
	}
}

func readTomlConfig(data []byte, args *lingualeoArgs) error {
	_, err := toml.Decode(string(data), &args)
	if err != nil {
		return err
	}
	return nil
}

func readYamlConfig(data []byte, args *lingualeoArgs) error {
	err := yaml.Unmarshal(data, args)
	if err != nil {
		return err
	}
	return nil
}

func readJSONConfig(data []byte, args *lingualeoArgs) error {
	err := json.Unmarshal(data, args)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFiles(filename *string) ([]string, error) {
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
	if filename != nil && len(*filename) > 0 {
		argsConfig, _ := filepath.Abs(*filename)
		if fileExists(argsConfig) {
			configs = append(configs, argsConfig)
		}
		configs = append(configs, argsConfig)
	}
	configs = unique(configs)
	return configs, nil
}

func readConfigs(filename *string) (*lingualeoArgs, error) {
	configs, err := getConfigFiles(filename)
	if err != nil {
		return nil, err
	}
	args := &lingualeoArgs{}
	for _, name := range configs {
		err = newConfigFileType(name).decodeFile(args)
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
