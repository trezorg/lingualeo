package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/alyu/configparser"
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
	configPtr := flag.String("c", "", "Config file")
	playerPtr := flag.String("m", "", "Media player for word pronounciation")
	forcePtr := flag.Bool("f", false, "Force add to lingualeo dictionary")
	addPtr := flag.Bool("a", false, "Add to lingualeo dictionary")
	soundPtr := flag.Bool("s", false, "Play words pronounciation")
	debugPtr := flag.Bool("d", false, "Debug mode")
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
		*debugPtr,
	}
}

func getUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
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

func readConfig(args *lingualeoArgs, filename string) error {
	config, err := configparser.Read(filename)
	if err != nil {
		return err
	}
	sections, err := config.AllSections()
	if err != nil {
		return err
	}
	options := sections[0].Options()
	for _, flag := range []string{"Email", "Password", "Player"} {
		setStringOption(args, flag, options)
	}
	for _, flag := range []string{"Force", "Add", "Sound", "Debug"} {
		err := setBoolOption(args, flag, options)
		if err != nil {
			return err
		}
	}
	args.Config = filename
	return nil
}

func readConfigs(filename string) (*lingualeoArgs, error) {
	home, err := getUserHome()
	if err != nil {
		return nil, err
	}
	homeConfig, _ := filepath.Abs(filepath.Join(home, defaultConfigFile))
	currentConfig, _ := filepath.Abs(defaultConfigFile)
	configs := []string{homeConfig, currentConfig}
	if len(filename) > 0 {
		argsConfig, _ := filepath.Abs(filename)
		configs = append(configs, argsConfig)
	}
	args := &lingualeoArgs{"", "", "", "", []string{}, []string{}, false, false, false, false}
	for _, name := range configs {
		readConfig(args, name)
	}
	return args, nil
}

func checkConfig(args *lingualeoArgs) error {
	if len(args.Config) > 0 {
		filename, _ := filepath.Abs(args.Config)
		if stat, err := os.Stat(filename); os.IsNotExist(err) || stat.IsDir() {
			return fmt.Errorf("There is no the config file or file is a directory: %s", filename)
		}
	}
	return nil
}

func checkArgs(args *lingualeoArgs) error {
	if len(args.Email) == 0 {
		return fmt.Errorf("No email argument has been supplied")
	}
	if len(args.Password) == 0 {
		return fmt.Errorf("No password argument has been supplied")
	}
	if len(args.Words) == 0 {
		return fmt.Errorf("No words to translate have been supplied")
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
	if configArgs.Debug {
		args.Debug = configArgs.Debug
	}
	return args
}
