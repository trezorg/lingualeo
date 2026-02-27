package translator

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/internal/slice"
	"github.com/trezorg/lingualeo/internal/validator"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"

	"gopkg.in/yaml.v3"
)

type (
	configType int
	decodeFunc func(data []byte, args *Lingualeo) error
)

const (
	yamlType = configType(1)
	jsonType = configType(2)
	tomlType = configType(3)
	// Default values for HTTP and retry settings
	defaultMaxIdleConns        = 10
	defaultMaxIdleConnsPerHost = 10
	defaultMaxRedirects        = 10
	defaultRetryMaxAttempts    = 3
	defaultRetryInitialWait    = 500 * time.Millisecond
	defaultRetryMaxWait        = 5 * time.Second
)

var decodeMapping = map[configType]decodeFunc{
	yamlType: readYamlConfig,
	jsonType: readJSONConfig,
	tomlType: readTomlConfig,
}

var (
	errNoWords                 = errors.New("there are no words to translate")
	errAddCustomTranslation    = errors.New("custom translation requires exactly one word")
	errConfigFileMissing       = errors.New("config file is missing or invalid")
	errEmailArgumentMissing    = errors.New("email argument is missing")
	errEmailInvalid            = errors.New("email argument is invalid")
	errPasswordArgumentMissing = errors.New("password argument is missing")

	// ErrHelpOrVersionShown is returned when --help or --version flag is passed.
	// The caller should treat this as a successful exit (os.Exit(0)).
	ErrHelpOrVersionShown = errors.New("help or version shown")
)

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

// isHelpOrVersion checks if the command line arguments contain help or version flags.
func isHelpOrVersion(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
			return true
		}
	}
	return false
}

func prepareArgs(version string) (Lingualeo, error) {
	args := Lingualeo{}
	translate := cli.StringSlice{}
	defaultCommand := buildDefaultCommand(&args, &translate)
	app := newLingualeoApp(version, &args, translate, defaultCommand)

	// Check if help/version flags are present - skip validation if so.
	// urfave/cli handles these flags and shows output, but we need to
	// prevent checkArgs() from running with empty Words.
	if isHelpOrVersion(os.Args) {
		_ = app.Run(os.Args)
		return args, ErrHelpOrVersionShown
	}

	return args, app.Run(os.Args)
}

func buildDefaultCommand(args *Lingualeo, translate *cli.StringSlice) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if c.NArg() == 0 {
			if err := cli.ShowAppHelp(c); err != nil {
				return fmt.Errorf("%w: %w", errNoWords, err)
			}
			return errNoWords
		}
		args.Words = slice.Unique(c.Args().Slice())
		args.Translation = slice.Unique(translate.Value())
		args.VisualiseType = *c.Generic("visualize-type").(*VisualiseType)
		if args.Add && len(args.Translation) > 0 && len(args.Words) > 1 {
			return errAddCustomTranslation
		}
		return nil
	}
}

func newLingualeoApp(version string, args *Lingualeo, translate cli.StringSlice, defaultCommand func(*cli.Context) error) *cli.App {
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
		"email": "email@gmail.com"
		"password": "password"
		"add": false
		"sound": true
		"player": "mplayer"
		"download": false
	}
	`
	app.Flags = buildLingualeoFlags(args)
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
	return app
}

func buildLingualeoFlags(args *Lingualeo) []cli.Flag {
	base := baseLingualeoFlags(args)
	base = append(base, httpAndRetryFlags(args)...)
	base = append(base, genericLingualeoFlags()...)
	return append(base, boolLingualeoFlags(args)...)
}

func baseLingualeoFlags(args *Lingualeo) []cli.Flag {
	return []cli.Flag{
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
		&cli.DurationFlag{
			Name:        "timeout",
			Aliases:     []string{"t"},
			Value:       defaultRequestTimeout,
			Usage:       "HTTP request timeout (e.g., 10s, 30s, 1m)",
			Destination: &args.RequestTimeout,
		},
	}
}

func httpAndRetryFlags(args *Lingualeo) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:        "max-idle-conns",
			Value:       defaultMaxIdleConns,
			Usage:       "Maximum number of idle HTTP connections",
			Destination: &args.MaxIdleConns,
		},
		&cli.IntFlag{
			Name:        "max-idle-conns-per-host",
			Value:       defaultMaxIdleConnsPerHost,
			Usage:       "Maximum idle HTTP connections per host",
			Destination: &args.MaxIdleConnsPerHost,
		},
		&cli.IntFlag{
			Name:        "max-redirects",
			Value:       defaultMaxRedirects,
			Usage:       "Maximum HTTP redirects to follow",
			Destination: &args.MaxRedirects,
		},
		&cli.IntFlag{
			Name:        "retry-max-attempts",
			Value:       defaultRetryMaxAttempts,
			Usage:       "Maximum retry attempts for failed requests",
			Destination: &args.RetryMaxAttempts,
		},
		&cli.DurationFlag{
			Name:        "retry-initial-wait",
			Value:       defaultRetryInitialWait,
			Usage:       "Initial wait between retry attempts",
			Destination: &args.RetryInitialWait,
		},
		&cli.DurationFlag{
			Name:        "retry-max-wait",
			Value:       defaultRetryMaxWait,
			Usage:       "Maximum wait between retry attempts",
			Destination: &args.RetryMaxWait,
		},
	}
}

func genericLingualeoFlags() []cli.Flag {
	return []cli.Flag{
		&cli.GenericFlag{
			Name:    "visualize-type",
			Aliases: []string{"vst"},
			Usage:   "Open picture either with default xdg-open or terminal graphic protocol. Allowed values: default, term",
			Value:   &VisualiseTypeDefault,
		},
	}
}

func boolLingualeoFlags(args *Lingualeo) []cli.Flag {
	return []cli.Flag{
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

func (l *Lingualeo) mergeConfigs(a *Lingualeo) {
	l.Email = mergeString(l.Email, a.Email)
	l.Password = mergeString(l.Password, a.Password)
	l.Player = mergeString(l.Player, a.Player)
	l.Add = mergeBool(l.Add, a.Add)
	l.Debug = mergeBool(l.Debug, a.Debug)
	l.Sound = mergeBool(l.Sound, a.Sound)
	l.Visualise = mergeBool(l.Visualise, a.Visualise)
	l.VisualiseType = VisualiseType(mergeString(string(l.VisualiseType), string(a.VisualiseType)))
	l.DownloadSoundFile = mergeBool(l.DownloadSoundFile, a.DownloadSoundFile)
	l.LogPrettyPrint = mergeBool(l.LogPrettyPrint, a.LogPrettyPrint)
	l.ReverseTranslate = mergeBool(l.ReverseTranslate, a.ReverseTranslate)
	l.LogLevel = mergeString(l.LogLevel, a.LogLevel)
	l.RequestTimeout = mergeDuration(l.RequestTimeout, a.RequestTimeout)
	// HTTP and retry settings
	l.MaxIdleConns = mergeInt(l.MaxIdleConns, a.MaxIdleConns)
	l.MaxIdleConnsPerHost = mergeInt(l.MaxIdleConnsPerHost, a.MaxIdleConnsPerHost)
	l.MaxRedirects = mergeInt(l.MaxRedirects, a.MaxRedirects)
	l.RetryMaxAttempts = mergeInt(l.RetryMaxAttempts, a.RetryMaxAttempts)
	l.RetryInitialWait = mergeDuration(l.RetryInitialWait, a.RetryInitialWait)
	l.RetryMaxWait = mergeDuration(l.RetryMaxWait, a.RetryMaxWait)
}

func mergeInt(dst, src int) int {
	if dst == 0 && src > 0 {
		return src
	}
	return dst
}

func mergeString(dst, src string) string {
	if len(dst) == 0 && len(src) > 0 {
		return src
	}
	return dst
}

func mergeBool(dst, src bool) bool {
	if !dst && src {
		return src
	}
	return dst
}

func mergeDuration(dst, src time.Duration) time.Duration {
	if dst == 0 && src > 0 {
		return src
	}
	return dst
}
