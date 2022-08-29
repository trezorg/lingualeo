package translator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/trezorg/lingualeo/pkg/utils"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"

	"gopkg.in/yaml.v2"
)

type configType int
type decodeFunc func(data []byte, args *Lingualeo) error

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

func (cf *configFile) decode(data []byte, args *Lingualeo) error {
	return decodeMapping[cf.getType()](data, args)
}

func (cf *configFile) decodeFile(args *Lingualeo) error {
	data, err := os.ReadFile(cf.filename)
	if err != nil {
		return err
	}
	return cf.decode(data, args)
}

func newConfigFile(filename string) *configFile {
	return &configFile{filename: filename}
}

func prepareCliArgs(version string) Lingualeo {

	args := Lingualeo{}

	var translate cli.StringSlice

	defaultCommand := func(c *cli.Context) error {
		if c.NArg() == 0 {
			err := cli.ShowAppHelp(c)
			if err != nil {
				return fmt.Errorf("there are no words to translate, %w", err)
			}
			return fmt.Errorf("there are no words to translate")
		}
		args.Words = utils.Unique(c.Args().Slice())
		args.Translation = translate.Value()
		if args.Add && len(args.Translation) > 0 && len(args.Words) > 1 {
			return fmt.Errorf("you should add only one word with custom transcation")
		}
		return nil
	}

	app := cli.NewApp()
	app.Version = version
	app.HideHelp = false
	app.HideVersion = false
	app.Authors = []*cli.Author{{
		Name:  "Igor Nemilentsev",
		Email: "trezorg@gmail.com",
	}}
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
	download = false

	Yaml format example:

	email: email@gmail.com
	password: password
	add: false
	sound: true
	player: mplayer
	download: false

	JSON format example:

	{
		"email": "email@gmail.com",
		"password": "password",
		"add": false,
		"sound": true,
		"player": "mplayer",
		"download": false,
	}
	`
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "email",
			Aliases:     []string{"e"},
			Value:       "",
			Usage:       "Lingualeo email",
			Destination: &args.Email,
		},
		&cli.StringFlag{
			Name:        "password",
			Aliases:     []string{"p"},
			Value:       "",
			Usage:       "Lingualeo password",
			Destination: &args.Password,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       "",
			Usage:       "Config file. Either in toml, yaml or json format",
			Destination: &args.Config,
		},
		&cli.StringFlag{
			Name:        "player",
			Aliases:     []string{"m"},
			Value:       "",
			Usage:       "Media player to pronounce words",
			Destination: &args.Player,
		},
		&cli.StringFlag{
			Name:        "log-level",
			Aliases:     []string{"l"},
			Value:       "INFO",
			Usage:       "Log level",
			Destination: &args.LogLevel,
		},
		&cli.BoolFlag{
			Name:        "sound",
			Aliases:     []string{"s"},
			Usage:       "Pronounce words",
			Destination: &args.Sound,
		},
		&cli.BoolFlag{
			Name:        "download",
			Aliases:     []string{"dl"},
			Usage:       "Download file to play sound. In case a player is not able to play url directly",
			Destination: &args.DownloadSoundFile,
		},
		&cli.BoolFlag{
			Name:        "debug",
			Aliases:     []string{"d"},
			Usage:       "Debug mode. Set DEBUG mode",
			Destination: &args.Debug,
		},
		&cli.BoolFlag{
			Name:        "log-pretty-print",
			Aliases:     []string{"lpr"},
			Usage:       "Log pretty print",
			Destination: &args.LogPrettyPrint,
		},
		&cli.BoolFlag{
			Name:        "reverse-translate",
			Aliases:     []string{"rt"},
			Usage:       "Reverse translate russian words",
			Destination: &args.ReverseTranslate,
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add to lingualeo dictionary",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "force",
					Aliases:     []string{"f"},
					Usage:       "Force add to lingualeo dictionary",
					Destination: &args.Force,
				},
				&cli.BoolFlag{
					Name:        "replace",
					Aliases:     []string{"r"},
					Usage:       "Custom translation. Replace word instead of adding",
					Destination: &args.TranslateReplace,
				},
				&cli.StringSliceFlag{
					Name:        "translate",
					Aliases:     []string{"t"},
					Usage:       "Custom translation: lingualeo add -t word1 -t word2 word",
					Destination: &translate,
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
		utils.FailIfError(err)
	}
	return args

}

func readTomlConfig(data []byte, args *Lingualeo) error {
	_, err := toml.Decode(string(data), args)
	if err != nil {
		return err
	}
	return nil
}

func readYamlConfig(data []byte, args *Lingualeo) error {
	err := yaml.Unmarshal(data, args)
	if err != nil {
		return err
	}
	return nil
}

func readJSONConfig(data []byte, args *Lingualeo) error {
	err := json.Unmarshal(data, args)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFiles(filename *string) ([]string, error) {
	home, err := utils.GetUserHome()
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
			if utils.FileExists(fullConfigFileName) {
				configs = append(configs, fullConfigFileName)
			}
		}
	}
	if filename != nil && len(*filename) > 0 {
		argsConfig, _ := filepath.Abs(*filename)
		if utils.FileExists(argsConfig) {
			configs = append(configs, argsConfig)
		}
	}
	configs = utils.Unique(configs)
	return configs, nil
}

func fromConfigs(filename *string) (*Lingualeo, error) {
	configs, err := getConfigFiles(filename)
	if err != nil {
		return nil, err
	}
	args := Lingualeo{}
	for _, name := range configs {
		err = newConfigFile(name).decodeFile(&args)
		if err != nil {
			return nil, err
		}
	}
	return &args, nil
}

func (args *Lingualeo) checkConfig() error {
	if len(args.Config) > 0 {
		filename, _ := filepath.Abs(args.Config)
		if !utils.FileExists(filename) {
			return fmt.Errorf("there is no the config file or file is a directory: %s", filename)
		}
	}
	return nil
}

func (args *Lingualeo) checkArgs() error {
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

func (args *Lingualeo) mergeConfigs(a *Lingualeo) {
	if len(args.Email) == 0 && len(a.Email) > 0 {
		args.Email = a.Email
	}
	if len(args.Password) == 0 && len(a.Password) > 0 {
		args.Password = a.Password
	}
	if len(args.Player) == 0 && len(a.Player) > 0 {
		args.Player = a.Player
	}
	if a.Force {
		args.Force = a.Force
	}
	if a.Add {
		args.Add = a.Add
	}
	if a.Debug {
		args.Debug = a.Debug
	}
	if a.Sound {
		args.Sound = a.Sound
	}
	if a.DownloadSoundFile {
		args.DownloadSoundFile = a.DownloadSoundFile
	}
	if a.LogPrettyPrint {
		args.LogPrettyPrint = a.LogPrettyPrint
	}
	if a.ReverseTranslate {
		args.ReverseTranslate = a.ReverseTranslate
	}
	if a.TranslateReplace {
		args.TranslateReplace = a.TranslateReplace
	}
	if len(args.LogLevel) == 0 && len(a.LogLevel) > 0 {
		args.LogLevel = a.LogLevel
	}
}
