package translator

import "time"

const defaultRequestTimeout = 10 * time.Second

var (
	defaultConfigFiles = []string{
		"lingualeo.toml",
		"lingualeo.yml",
		"lingualeo.yaml",
		"lingualeo.json",
	}
)
