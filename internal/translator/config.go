package translator

import (
	"cmp"
	"time"

	"github.com/trezorg/lingualeo/internal/api"
)

// Config holds all serializable configuration for Lingualeo.
type Config struct {
	// Authentication
	Email    string `yaml:"email" json:"email" toml:"email"`
	Password string `yaml:"password" json:"password" toml:"password"` //nolint:gosec // credential field

	// Media player
	Player                string        `yaml:"player" json:"player" toml:"player"`
	PlayerShutdownTimeout time.Duration `yaml:"player_shutdown_timeout" json:"player_shutdown_timeout" toml:"player_shutdown_timeout"`

	// Logging
	LogLevel       string `yaml:"log_level" json:"log_level" toml:"log_level"`
	LogPrettyPrint bool   `yaml:"log_pretty_print" json:"log_pretty_print" toml:"log_pretty_print"`
	Debug          bool   `yaml:"debug" json:"debug" toml:"debug"`

	// Behavior toggles
	Add               bool          `yaml:"add" json:"add" toml:"add"`
	Sound             bool          `yaml:"sound" json:"sound" toml:"sound"`
	Visualise         bool          `yaml:"visualize" json:"visualize" toml:"visualize"`
	VisualiseType     VisualiseType `yaml:"visualize_type" json:"visualize_type" toml:"visualize_type"`
	DownloadSoundFile bool          `yaml:"download" json:"download" toml:"download"`
	ReverseTranslate  bool          `yaml:"reverse_translate" json:"reverse_translate" toml:"reverse_translate"`
	PromptPassword    bool          `yaml:"prompt_password" json:"prompt_password" toml:"prompt_password"`

	// Concurrency
	Workers int `yaml:"workers" json:"workers" toml:"workers"`

	// HTTP settings
	RequestTimeout      time.Duration `yaml:"request_timeout" json:"request_timeout" toml:"request_timeout"`
	MaxIdleConns        int           `yaml:"max_idle_conns" json:"max_idle_conns" toml:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" json:"max_idle_conns_per_host" toml:"max_idle_conns_per_host"`
	MaxRedirects        int           `yaml:"max_redirects" json:"max_redirects" toml:"max_redirects"`
	RetryMaxAttempts    int           `yaml:"retry_max_attempts" json:"retry_max_attempts" toml:"retry_max_attempts"`
	RetryInitialWait    time.Duration `yaml:"retry_initial_wait" json:"retry_initial_wait" toml:"retry_initial_wait"`
	RetryMaxWait        time.Duration `yaml:"retry_max_wait" json:"retry_max_wait" toml:"retry_max_wait"`
}

// APIClientConfig converts Config to api.Config for the HTTP client.
func (c *Config) APIClientConfig() api.Config {
	timeout := c.RequestTimeout
	if timeout == 0 {
		timeout = api.DefaultConfig().Timeout
	}
	return api.Config{
		Timeout:             timeout,
		MaxRedirects:        c.MaxRedirects,
		MaxIdleConns:        c.MaxIdleConns,
		MaxIdleConnsPerHost: c.MaxIdleConnsPerHost,
		Retry: api.RetryConfig{
			MaxAttempts: c.RetryMaxAttempts,
			InitialWait: c.RetryInitialWait,
			MaxWait:     c.RetryMaxWait,
		},
	}
}

// Merge merges non-zero values from src into dst, preserving dst's non-zero values.
func (c *Config) Merge(src *Config) {
	c.Email = cmp.Or(c.Email, src.Email)
	c.Password = cmp.Or(c.Password, src.Password)
	c.Player = cmp.Or(c.Player, src.Player)
	c.Add = cmp.Or(c.Add, src.Add)
	c.Debug = cmp.Or(c.Debug, src.Debug)
	c.Sound = cmp.Or(c.Sound, src.Sound)
	c.Visualise = cmp.Or(c.Visualise, src.Visualise)
	c.VisualiseType = VisualiseType(cmp.Or(string(c.VisualiseType), string(src.VisualiseType)))
	c.DownloadSoundFile = cmp.Or(c.DownloadSoundFile, src.DownloadSoundFile)
	c.LogPrettyPrint = cmp.Or(c.LogPrettyPrint, src.LogPrettyPrint)
	c.ReverseTranslate = cmp.Or(c.ReverseTranslate, src.ReverseTranslate)
	c.PromptPassword = cmp.Or(c.PromptPassword, src.PromptPassword)
	c.LogLevel = cmp.Or(c.LogLevel, src.LogLevel)
	c.Workers = cmp.Or(c.Workers, src.Workers)
	c.RequestTimeout = cmp.Or(c.RequestTimeout, src.RequestTimeout)
	c.PlayerShutdownTimeout = cmp.Or(c.PlayerShutdownTimeout, src.PlayerShutdownTimeout)
	c.MaxIdleConns = cmp.Or(c.MaxIdleConns, src.MaxIdleConns)
	c.MaxIdleConnsPerHost = cmp.Or(c.MaxIdleConnsPerHost, src.MaxIdleConnsPerHost)
	c.MaxRedirects = cmp.Or(c.MaxRedirects, src.MaxRedirects)
	c.RetryMaxAttempts = cmp.Or(c.RetryMaxAttempts, src.RetryMaxAttempts)
	c.RetryInitialWait = cmp.Or(c.RetryInitialWait, src.RetryInitialWait)
	c.RetryMaxWait = cmp.Or(c.RetryMaxWait, src.RetryMaxWait)
}
