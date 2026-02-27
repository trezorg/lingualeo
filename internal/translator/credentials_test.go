package translator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCredentialFlagsUseEnvVars(t *testing.T) {
	args := Lingualeo{}
	flags := baseLingualeoFlags(&args)

	var emailFlag *cli.StringFlag
	var passwordFlag *cli.StringFlag
	for _, flag := range flags {
		stringFlag, ok := flag.(*cli.StringFlag)
		if !ok {
			continue
		}
		switch stringFlag.Name {
		case "email":
			emailFlag = stringFlag
		case "password":
			passwordFlag = stringFlag
		}
	}

	require.NotNil(t, emailFlag)
	require.NotNil(t, passwordFlag)
	assert.Equal(t, []string{"LINGUALEO_EMAIL"}, emailFlag.EnvVars)
	assert.Equal(t, []string{"LINGUALEO_PASSWORD"}, passwordFlag.EnvVars)
}

func TestPromptPasswordIfNeeded(t *testing.T) {
	originalPrompt := passwordPrompt
	t.Cleanup(func() {
		passwordPrompt = originalPrompt
	})

	tests := []struct {
		name               string
		args               Lingualeo
		promptResult       string
		promptErr          error
		expectPromptCalled bool
		expectedPassword   string
		expectErr          bool
	}{
		{
			name: "prompts when enabled and password missing",
			args: Lingualeo{
				PromptPassword: true,
			},
			promptResult:       "prompted-secret",
			expectPromptCalled: true,
			expectedPassword:   "prompted-secret",
			expectErr:          false,
		},
		{
			name: "skips prompt when password already set",
			args: Lingualeo{
				PromptPassword: true,
				Password:       "from-config",
			},
			expectPromptCalled: false,
			expectedPassword:   "from-config",
			expectErr:          false,
		},
		{
			name: "skips prompt when option is disabled",
			args: Lingualeo{
				PromptPassword: false,
			},
			expectPromptCalled: false,
			expectedPassword:   "",
			expectErr:          false,
		},
		{
			name: "returns prompt error",
			args: Lingualeo{
				PromptPassword: true,
			},
			promptErr:          errors.New("prompt failed"),
			expectPromptCalled: true,
			expectedPassword:   "",
			expectErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptCalled := false
			passwordPrompt = func() (string, error) {
				promptCalled = true
				return tt.promptResult, tt.promptErr
			}

			err := tt.args.promptPasswordIfNeeded()

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectPromptCalled, promptCalled)
			assert.Equal(t, tt.expectedPassword, tt.args.Password)
		})
	}
}
