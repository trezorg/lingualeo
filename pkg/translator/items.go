package translator

import "github.com/trezorg/lingualeo/pkg/api"

type Lingualeo struct {
	Email             string `yaml:"email" json:"email" toml:"email"`
	Password          string `yaml:"password" json:"password" toml:"password"`
	Config            string
	Player            string `yaml:"player" json:"player" toml:"player"`
	Words             []string
	Translation       []string
	Force             bool   `yaml:"force" json:"force" toml:"force"`
	Add               bool   `yaml:"add" json:"add" toml:"add"`
	TranslateReplace  bool   `yaml:"translate_replace" json:"translate_replace" toml:"translate_replace"`
	Sound             bool   `yaml:"sound" json:"sound" toml:"sound"`
	Debug             bool   `yaml:"debug" json:"debug" toml:"debug"`
	DownloadSoundFile bool   `yaml:"download" json:"download" toml:"download"`
	LogLevel          string `yaml:"log_level" json:"log_level" toml:"log_level"`
	LogPrettyPrint    bool   `yaml:"log_pretty_print" json:"log_pretty_print" toml:"log_pretty_print"`
	ReverseTranslate  bool   `yaml:"reverse_translate" json:"reverse_translate" toml:"reverse_translate"`
	API               api.Translator
}
