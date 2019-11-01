package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli"

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

type configFile struct {
	filename string
}

func (cf *configFile) getType() configType {
	ext := filepath.Ext(cf.filename)
	if ext == ".yml" || ext == ".yaml" {
		return yamlType
	}
	if ext == ".json" {
		return jsonType
	}
	return tomlType
}

func (cf *configFile) decode(data []byte, args *lingualeoArgs) error {
	return decodeMapping[cf.getType()](data, args)
}

func (cf *configFile) decodeFile(args *lingualeoArgs) error {
	data, err := ioutil.ReadFile(cf.filename)
	if err != nil {
		return err
	}
	return cf.decode(data, args)
}

func newConfigFile(filename string) *configFile {
	return &configFile{filename: filename}
}

func prepareCliArgs() lingualeoArgs {

	args := lingualeoArgs{}

	defaultCommand := func(c *cli.Context) error {
		if c.NArg() == 0 {
			err := cli.ShowAppHelp(c)
			if err != nil {
				return fmt.Errorf("there are no words to translate, %w", err)
			}
			return fmt.Errorf("there are no words to translate")
		}
		args.Words = unique(c.Args())
		if args.Add && len(args.Translate) > 0 && len(args.Words) > 1 {
			return fmt.Errorf("you should add only one word with custom transcation")
		}
		return nil
	}

	app := cli.NewApp()
	app.Version = "0.0.1"
	app.HideHelp = true
	app.HideVersion = true
	app.Author = "Igor Nemilentsev"
	app.Email = "trezorg@gmail.com"
	app.Usage = "Lingualeo API console helper"
	app.EnableBashCompletion = true
	app.ArgsUsage = "Multiple words can be supplied"
	app.Action = defaultCommand
	app.Description = `
	It is possible to use config file to set predefined parameters  	
	Default config files are: ~/lingualeo.[toml|yml|yaml|json]

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
	`
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "email, e",
			Value:       "",
			Usage:       "Lingualeo email",
			Destination: &args.Email,
		},
		cli.StringFlag{
			Name:        "password, p",
			Value:       "",
			Usage:       "Lingualeo password",
			Destination: &args.Password,
		},
		cli.StringFlag{
			Name:        "config, c",
			Value:       "",
			Usage:       "Config file. Either in toml, yaml or json format",
			Destination: &args.Config,
		},
		cli.StringFlag{
			Name:        "player, m",
			Value:       "",
			Usage:       "Media player for word pronunciation",
			Destination: &args.Player,
		},
		cli.StringFlag{
			Name:        "log-level, l",
			Value:       "INFO",
			Usage:       "Log level",
			Destination: &args.LogLevel,
		},
		cli.BoolFlag{
			Name:        "sound, s",
			Usage:       "Play words pronunciation",
			Destination: &args.Sound,
		},
		cli.BoolFlag{
			Name:        "download, dl",
			Usage:       "Download file to play sound",
			Destination: &args.DownloadSoundFile,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "Debug mode. Set DEBUG mode",
			Destination: &args.Debug,
		},
		cli.BoolFlag{
			Name:        "log-pretty-print, lpr",
			Usage:       "Log pretty print",
			Destination: &args.LogPrettyPrint,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add to lingualeo dictionary",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "force, f",
					Usage:       "Force add to lingualeo dictionary",
					Destination: &args.Force,
				},
				cli.BoolFlag{
					Name:        "replace, r",
					Usage:       "Custom translation. Replace word instead of to add",
					Destination: &args.TranslateReplace,
				},
				cli.StringSliceFlag{
					Name:  "translate, t",
					Usage: "Custom translation: lingualeo add -t word1 -t word2 word",
					Value: &args.Translate,
				},
			},
			Action: func(c *cli.Context) error {
				args.Add = true
				return defaultCommand(c)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		if strings.Contains(err.Error(), "help requested") {
			os.Exit(0)
		}
		failIfError(err)
	}
	return args

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
		err = newConfigFile(name).decodeFile(args)
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
	if configArgs.DownloadSoundFile {
		args.DownloadSoundFile = configArgs.DownloadSoundFile
	}
	if configArgs.LogPrettyPrint {
		args.LogPrettyPrint = configArgs.LogPrettyPrint
	}
	if configArgs.TranslateReplace {
		args.TranslateReplace = configArgs.TranslateReplace
	}
	if len(args.LogLevel) == 0 && len(configArgs.LogLevel) > 0 {
		args.LogLevel = configArgs.LogLevel
	}
	return args
}
