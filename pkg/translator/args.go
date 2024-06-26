package translator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/internal/slice"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"

	"gopkg.in/yaml.v2"
)

type (
	configType int
	decodeFunc func(data []byte, args *Lingualeo) error
)

const (
	yamlType = configType(1)
	jsonType = configType(2)
	tomlType = configType(3)
)

var decodeMapping = map[configType]decodeFunc{
	yamlType: readYamlConfig,
	jsonType: readJSONConfig,
	tomlType: readTomlConfig,
}

type configFile struct {
	filename string
}

func userHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func (cf *configFile) filenameType() configType {
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
	return decodeMapping[cf.filenameType()](data, args)
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

func prepareArgs(version string) (Lingualeo, error) {
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
		args.Words = slice.Unique(c.Args().Slice())
		args.Translation = slice.Unique(translate.Value())
		args.VisualiseType = *c.Generic("visualize-type").(*VisualiseType)
		if args.Add && len(args.Translation) > 0 && len(args.Words) > 1 {
			return fmt.Errorf("you should add only one word with custom translation")
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
			Name:        "visualize",
			Aliases:     []string{"vs"},
			Usage:       "Open translate pictures",
			Destination: &args.Visualise,
		},
		&cli.GenericFlag{
			Name:    "visualize-type",
			Aliases: []string{"vst"},
			Usage:   "Open picture either with default xgd-open or terminal graphic protocol (kitty, iterm, sizel)",
			Value:   &VisualiseTypeDefault,
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
				&cli.StringSliceFlag{
					Name:        "translate",
					Aliases:     []string{"t"},
					Usage:       "Custom translation: lingualeo add -t word1 -t word2 word",
					Destination: &translate,
					Required:    true,
				},
			},
			Action: func(c *cli.Context) error {
				args.Add = true
				return defaultCommand(c)
			},
		},
	}

	return args, app.Run(os.Args)
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

func configFiles(filename *string) ([]string, error) {
	home, err := userHome()
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
			if files.Exists(fullConfigFileName) {
				configs = append(configs, fullConfigFileName)
			}
		}
	}
	if filename != nil && len(*filename) > 0 {
		argsConfig, _ := filepath.Abs(*filename)
		if files.Exists(argsConfig) {
			configs = append(configs, argsConfig)
		}
	}
	configs = slice.Unique(configs)
	return configs, nil
}

func fromConfigs(filename *string) (*Lingualeo, error) {
	configs, err := configFiles(filename)
	if err != nil {
		return nil, err
	}
	args := Lingualeo{}
	for _, name := range configs {
		if err = newConfigFile(name).decodeFile(&args); err != nil {
			return nil, err
		}
	}
	return &args, nil
}

func (l *Lingualeo) checkConfig() error {
	if len(l.Config) > 0 {
		filename, _ := filepath.Abs(l.Config)
		if !files.Exists(filename) {
			return fmt.Errorf("there is no the config file or file is a directory: %s", filename)
		}
	}
	return nil
}

func (l *Lingualeo) checkArgs() error {
	if len(l.Email) == 0 {
		return fmt.Errorf("mo email argument has been supplied")
	}
	if len(l.Password) == 0 {
		return fmt.Errorf("no password argument has been supplied")
	}
	if len(l.Words) == 0 {
		return fmt.Errorf("no words to translate have been supplied")
	}
	return nil
}

func (l *Lingualeo) mergeConfigs(a *Lingualeo) {
	if len(l.Email) == 0 && len(a.Email) > 0 {
		l.Email = a.Email
	}
	if len(l.Password) == 0 && len(a.Password) > 0 {
		l.Password = a.Password
	}
	if len(l.Player) == 0 && len(a.Player) > 0 {
		l.Player = a.Player
	}
	if a.Add {
		l.Add = a.Add
	}
	if a.Debug {
		l.Debug = a.Debug
	}
	if a.Sound {
		l.Sound = a.Sound
	}
	if a.Visualise {
		l.Visualise = a.Visualise
	}
	if len(a.VisualiseType) > 0 {
		l.VisualiseType = a.VisualiseType
	}
	if a.DownloadSoundFile {
		l.DownloadSoundFile = a.DownloadSoundFile
	}
	if a.LogPrettyPrint {
		l.LogPrettyPrint = a.LogPrettyPrint
	}
	if a.ReverseTranslate {
		l.ReverseTranslate = a.ReverseTranslate
	}
	if len(l.LogLevel) == 0 && len(a.LogLevel) > 0 {
		l.LogLevel = a.LogLevel
	}
}
