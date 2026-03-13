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

const defaultLogLevel = "INFO"

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

func (c *Config) ApplyDefaults() {
	defaults := api.DefaultConfig()
	c.LogLevel = cmp.Or(c.LogLevel, defaultLogLevel)
	c.Workers = cmp.Or(c.Workers, defaultWorkers)
	c.VisualiseType = VisualiseType(cmp.Or(string(c.VisualiseType), string(VisualiseTypeDefault)))
	c.RequestTimeout = cmp.Or(c.RequestTimeout, defaults.Timeout)
	c.MaxIdleConns = cmp.Or(c.MaxIdleConns, defaults.MaxIdleConns)
	c.MaxIdleConnsPerHost = cmp.Or(c.MaxIdleConnsPerHost, defaults.MaxIdleConnsPerHost)
	c.MaxRedirects = cmp.Or(c.MaxRedirects, defaults.MaxRedirects)
	c.RetryMaxAttempts = cmp.Or(c.RetryMaxAttempts, defaults.Retry.MaxAttempts)
	c.RetryInitialWait = cmp.Or(c.RetryInitialWait, defaults.Retry.InitialWait)
	c.RetryMaxWait = cmp.Or(c.RetryMaxWait, defaults.Retry.MaxWait)
}
