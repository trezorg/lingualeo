package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckEitherCommandIsAvailableOnSystem(t *testing.T) {
	assert.True(t, isCommandAvailable("bash -c 'oops'"))
	assert.True(t, isCommandAvailable("bash"))
	assert.False(t, isCommandAvailable("xxxxxxxxxxx"))
}

func TestRussianWord(t *testing.T) {
	assert.True(t, isRussianWord("гном"))
	assert.True(t, isRussianWord("гном1"))
	assert.False(t, isRussianWord("test"))
}

func TestCheckArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    Lingualeo
		wantErr bool
	}{
		{
			name: "valid args",
			args: Lingualeo{
				Email:    "user@example.com",
				Password: "password",
				Words:    []string{"test"},
			},
			wantErr: false,
		},
		{
			name: "missing email",
			args: Lingualeo{
				Password: "password",
				Words:    []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "invalid email - no @",
			args: Lingualeo{
				Email:    "userexample.com",
				Password: "password",
				Words:    []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "invalid email - no domain",
			args: Lingualeo{
				Email:    "user@",
				Password: "password",
				Words:    []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "missing password",
			args: Lingualeo{
				Email: "user@example.com",
				Words: []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "missing words",
			args: Lingualeo{
				Email:    "user@example.com",
				Password: "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.checkArgs()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
