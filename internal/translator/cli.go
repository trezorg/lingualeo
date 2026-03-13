package translator

import (
	"fmt"
	"os"

	"github.com/trezorg/lingualeo/internal/slice"

	"github.com/urfave/cli/v2"
)

func isHelpOrVersion(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
			return true
		}
	}

	return false
}

func prepareTranslator(version string) (Lingualeo, error) {
	translator := Lingualeo{}
	translate := cli.StringSlice{}
	defaultCommand := buildDefaultCommand(&translator, &translate)

	if isHelpOrVersion(os.Args) {
		app := newLingualeoApp(version, &translator, &translate, defaultCommand)
		_ = app.Run(os.Args)
		return translator, ErrHelpOrVersionShown
	}

	translator.ConfigPath = explicitConfigPath(os.Args)
	if err := translator.checkConfig(); err != nil {
		return translator, err
	}

	config, err := loadConfig(translator.ConfigPath)
	if err != nil {
		return translator, err
	}
	translator.Config = config

	cliApp := newLingualeoApp(version, &translator, &translate, defaultCommand)

	return translator, cliApp.Run(os.Args)
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

func newLingualeoApp(version string, args *Lingualeo, translate *cli.StringSlice, defaultCommand func(*cli.Context) error) *cli.App {
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
	Credentials can also be provided via LINGUALEO_EMAIL and LINGUALEO_PASSWORD env vars

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
					Destination: translate,
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
	base = append(base, genericLingualeoFlags(args)...)

	return append(base, boolLingualeoFlags(args)...)
}

func baseLingualeoFlags(args *Lingualeo) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "email",
			Aliases:     []string{"e"},
			Value:       args.Email,
			Usage:       "Lingualeo email",
			EnvVars:     []string{"LINGUALEO_EMAIL"},
			Destination: &args.Email,
		},
		&cli.StringFlag{
			Name:        "password",
			Aliases:     []string{"p"},
			Value:       args.Password,
			Usage:       "Lingualeo password",
			EnvVars:     []string{"LINGUALEO_PASSWORD"},
			Destination: &args.Password,
		},
		&cli.BoolFlag{
			Name:        "prompt-password",
			Usage:       "Prompt for Lingualeo password without echo",
			Value:       args.PromptPassword,
			Destination: &args.PromptPassword,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       args.ConfigPath,
			Usage:       "Config file. Either in toml, yaml or json format",
			Destination: &args.ConfigPath,
		},
		&cli.StringFlag{
			Name:        "player",
			Aliases:     []string{"m"},
			Value:       args.Player,
			Usage:       "Media player to pronounce words",
			Destination: &args.Player,
		},
		&cli.StringFlag{
			Name:        "log-level",
			Aliases:     []string{"l"},
			Value:       args.LogLevel,
			Usage:       "Log level",
			Destination: &args.LogLevel,
		},
		&cli.DurationFlag{
			Name:        "timeout",
			Aliases:     []string{"t"},
			Value:       args.RequestTimeout,
			Usage:       "HTTP request timeout (e.g., 10s, 30s, 1m)",
			Destination: &args.RequestTimeout,
		},
		&cli.IntFlag{
			Name:        "workers",
			Value:       args.Workers,
			Usage:       "Maximum number of concurrent workers for translate/add pipelines",
			Destination: &args.Workers,
		},
	}
}

func httpAndRetryFlags(args *Lingualeo) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:        "max-idle-conns",
			Value:       args.MaxIdleConns,
			Usage:       "Maximum number of idle HTTP connections",
			Destination: &args.MaxIdleConns,
		},
		&cli.IntFlag{
			Name:        "max-idle-conns-per-host",
			Value:       args.MaxIdleConnsPerHost,
			Usage:       "Maximum idle HTTP connections per host",
			Destination: &args.MaxIdleConnsPerHost,
		},
		&cli.IntFlag{
			Name:        "max-redirects",
			Value:       args.MaxRedirects,
			Usage:       "Maximum HTTP redirects to follow",
			Destination: &args.MaxRedirects,
		},
		&cli.IntFlag{
			Name:        "retry-max-attempts",
			Value:       args.RetryMaxAttempts,
			Usage:       "Maximum retry attempts for failed requests",
			Destination: &args.RetryMaxAttempts,
		},
		&cli.DurationFlag{
			Name:        "retry-initial-wait",
			Value:       args.RetryInitialWait,
			Usage:       "Initial wait between retry attempts",
			Destination: &args.RetryInitialWait,
		},
		&cli.DurationFlag{
			Name:        "retry-max-wait",
			Value:       args.RetryMaxWait,
			Usage:       "Maximum wait between retry attempts",
			Destination: &args.RetryMaxWait,
		},
	}
}

func genericLingualeoFlags(args *Lingualeo) []cli.Flag {
	return []cli.Flag{
		&cli.GenericFlag{
			Name:    "visualize-type",
			Aliases: []string{"vst"},
			Usage:   "Open picture either with default xdg-open or terminal graphic protocol. Allowed values: default, term",
			Value:   new(args.VisualiseType),
		},
	}
}

func boolLingualeoFlags(args *Lingualeo) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "sound",
			Aliases:     []string{"s"},
			Usage:       "Pronounce words",
			Value:       args.Sound,
			Destination: &args.Sound,
		},
		&cli.BoolFlag{
			Name:        "visualize",
			Aliases:     []string{"vs"},
			Usage:       "Open translate pictures",
			Value:       args.Visualise,
			Destination: &args.Visualise,
		},
		&cli.BoolFlag{
			Name:        "download",
			Aliases:     []string{"dl"},
			Usage:       "Download file to play sound. In case a player is not able to play url directly",
			Value:       args.DownloadSoundFile,
			Destination: &args.DownloadSoundFile,
		},
		&cli.BoolFlag{
			Name:        "debug",
			Aliases:     []string{"d"},
			Usage:       "Debug mode. Set DEBUG mode",
			Value:       args.Debug,
			Destination: &args.Debug,
		},
		&cli.BoolFlag{
			Name:        "log-pretty-print",
			Aliases:     []string{"lpr"},
			Usage:       "Log pretty print",
			Value:       args.LogPrettyPrint,
			Destination: &args.LogPrettyPrint,
		},
		&cli.BoolFlag{
			Name:        "reverse-translate",
			Aliases:     []string{"rt"},
			Usage:       "Reverse translate russian words",
			Value:       args.ReverseTranslate,
			Destination: &args.ReverseTranslate,
		},
	}
}
