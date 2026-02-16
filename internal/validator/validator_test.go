package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://example.com/path",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			url:     "http://example.com/path",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: false,
		},
		{
			name:    "invalid scheme - ftp",
			url:     "ftp://example.com/file",
			wantErr: true,
		},
		{
			name:    "file scheme not allowed",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "javascript scheme not allowed",
			url:     "javascript:alert(1)",
			wantErr: true,
		},
		{
			name:    "empty host",
			url:     "http:///path",
			wantErr: true,
		},
		{
			name:    "no scheme",
			url:     "example.com/path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: false,
		},
		{
			name:    "missing @",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "@ at start",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "@ at end",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "domain without dot is valid per RFC",
			email:   "user@localhost",
			wantErr: false,
		},
		{
			name:    "contains space",
			email:   "user @example.com",
			wantErr: true,
		},
		{
			name:    "space after @",
			email:   "user@ example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
