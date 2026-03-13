package translator

import (
	"encoding/json/v2"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/trezorg/lingualeo/internal/files"
	"github.com/trezorg/lingualeo/internal/slice"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type (
	configType int
	decodeFunc func(data []byte, cfg *Config) error
)

const (
	yamlType = configType(1)
	jsonType = configType(2)
	tomlType = configType(3)
)

var decodeMapping = map[configType]decodeFunc{
	yamlType: readYAMLConfig,
	jsonType: readJSONConfig,
	tomlType: readTOMLConfig,
}

var lookupUserHome = currentUserHome

type configFile struct {
	filename string
}

func currentUserHome() (string, error) {
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

func (cf *configFile) decode(data []byte, cfg *Config) error {
	return decodeMapping[cf.filenameType()](data, cfg)
}

func (cf *configFile) decodeFile(cfg *Config) error {
	data, err := os.ReadFile(cf.filename)
	if err != nil {
		return err
	}

	return cf.decode(data, cfg)
}

func newConfigFile(filename string) *configFile {
	return &configFile{filename: filename}
}

func readTOMLConfig(data []byte, cfg *Config) error {
	_, err := toml.Decode(string(data), cfg)
	if err != nil {
		return err
	}

	return nil
}

func readYAMLConfig(data []byte, cfg *Config) error {
	err := yaml.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	return nil
}

func readJSONConfig(data []byte, cfg *Config) error {
	err := json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	return nil
}

func explicitConfigPath(args []string) string {
	for i := range args {
		arg := args[i]
		switch {
		case arg == "--config" || arg == "-c":
			if i+1 >= len(args) {
				return ""
			}
			return args[i+1]
		case strings.HasPrefix(arg, "--config="):
			return strings.TrimPrefix(arg, "--config=")
		case strings.HasPrefix(arg, "-c="):
			return strings.TrimPrefix(arg, "-c=")
		}
	}

	return ""
}

func configFiles(filename string) ([]string, error) {
	home, err := lookupUserHome()
	if err != nil {
		return nil, err
	}

	configs := make([]string, 0, len(defaultConfigFiles)*2+1)
	for _, configFilename := range defaultConfigFiles {
		homeConfigFile, homeErr := filepath.Abs(filepath.Join(home, configFilename))
		if homeErr != nil {
			return nil, fmt.Errorf("resolve home config path: %w", homeErr)
		}
		currentConfigFile, currentErr := filepath.Abs(configFilename)
		if currentErr != nil {
			return nil, fmt.Errorf("resolve current config path: %w", currentErr)
		}

		for _, fullConfigFileName := range [2]string{homeConfigFile, currentConfigFile} {
			if files.Exists(fullConfigFileName) {
				configs = append(configs, fullConfigFileName)
			}
		}
	}

	if filename != "" {
		argsConfig, absErr := filepath.Abs(filename)
		if absErr != nil {
			return nil, fmt.Errorf("resolve explicit config path: %w", absErr)
		}
		if files.Exists(argsConfig) {
			configs = append(configs, argsConfig)
		}
	}

	return slice.Unique(configs), nil
}

func loadConfig(filename string) (Config, error) {
	configs, err := configFiles(filename)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{}
	for _, name := range configs {
		if err = newConfigFile(name).decodeFile(&cfg); err != nil {
			return Config{}, err
		}
	}

	cfg.ApplyDefaults()

	return cfg, nil
}
