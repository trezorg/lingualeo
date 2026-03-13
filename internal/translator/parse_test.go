package translator

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseUsesConfigValuesWhenFlagsAreOmitted(t *testing.T) {
	useTempHome(t)
	t.Chdir(t.TempDir())
	writeConfig(t, "lingualeo.toml", `
email = "config@example.com"
password = "secret"
log_level = "ERROR"
workers = 9
request_timeout = "42s"
`)

	withArgs(t, []string{"lingualeo", "hello"})

	client, err := Parse("test")
	require.NoError(t, err)
	require.Equal(t, "config@example.com", client.Email)
	require.Equal(t, "secret", client.Password)
	require.Equal(t, "ERROR", client.LogLevel)
	require.Equal(t, 9, client.Workers)
	require.Equal(t, 42*time.Second, client.RequestTimeout)
}

func TestParseCLIOverridesConfigValues(t *testing.T) {
	useTempHome(t)
	t.Chdir(t.TempDir())
	writeConfig(t, "lingualeo.toml", `
email = "config@example.com"
password = "secret"
log_level = "ERROR"
workers = 9
request_timeout = "42s"
`)

	withArgs(t, []string{"lingualeo", "--log-level", "DEBUG", "--workers", "2", "--timeout", "5s", "hello"})

	client, err := Parse("test")
	require.NoError(t, err)
	require.Equal(t, "DEBUG", client.LogLevel)
	require.Equal(t, 2, client.Workers)
	require.Equal(t, 5*time.Second, client.RequestTimeout)
}

func TestParseEnvOverridesConfigCredentials(t *testing.T) {
	useTempHome(t)
	t.Chdir(t.TempDir())
	writeConfig(t, "lingualeo.toml", `
email = "config@example.com"
password = "config-secret"
`)
	t.Setenv("LINGUALEO_EMAIL", "env@example.com")
	t.Setenv("LINGUALEO_PASSWORD", "env-secret")

	withArgs(t, []string{"lingualeo", "hello"})

	client, err := Parse("test")
	require.NoError(t, err)
	require.Equal(t, "env@example.com", client.Email)
	require.Equal(t, "env-secret", client.Password)
}

func TestParseExplicitConfigOverridesDiscoveredConfig(t *testing.T) {
	useTempHome(t)
	t.Chdir(t.TempDir())
	writeConfig(t, "lingualeo.toml", `
email = "discovered@example.com"
password = "discovered-secret"
log_level = "INFO"
`)
	explicitPath := writeConfig(t, "custom.toml", `
email = "explicit@example.com"
password = "explicit-secret"
log_level = "WARN"
`)

	withArgs(t, []string{"lingualeo", "--config", explicitPath, "hello"})

	client, err := Parse("test")
	require.NoError(t, err)
	require.Equal(t, "explicit@example.com", client.Email)
	require.Equal(t, "explicit-secret", client.Password)
	require.Equal(t, "WARN", client.LogLevel)
}

func TestConfigFilesIncludeExplicitConfigLast(t *testing.T) {
	t.Chdir(t.TempDir())
	homeDir := useTempHome(t)

	homeConfig := writeConfigAt(t, filepath.Join(homeDir, "lingualeo.toml"), "email = \"home@example.com\"\n")
	currentConfig := writeConfig(t, "lingualeo.toml", "email = \"cwd@example.com\"\n")
	explicitConfig := writeConfig(t, "custom.toml", "email = \"explicit@example.com\"\n")

	configs, err := configFiles(explicitConfig)
	require.NoError(t, err)
	require.Equal(t, []string{homeConfig, currentConfig, explicitConfig}, configs)
}

func withArgs(t *testing.T, args []string) {
	t.Helper()

	originalArgs := os.Args
	os.Args = args
	t.Cleanup(func() {
		os.Args = originalArgs
	})
}

func writeConfig(t *testing.T, filename string, content string) string {
	t.Helper()

	return writeConfigAt(t, filepath.Join(".", filename), content)
}

func writeConfigAt(t *testing.T, filename string, content string) string {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
	require.NoError(t, os.WriteFile(filename, []byte(content), 0o600))
	path, err := filepath.Abs(filename)
	require.NoError(t, err)

	return path
}

func useTempHome(t *testing.T) string {
	t.Helper()

	homeDir := t.TempDir()
	lookupUserHome = func() (string, error) {
		return homeDir, nil
	}
	t.Cleanup(func() {
		lookupUserHome = currentUserHome
	})

	return homeDir
}
