package validator

import (
	"errors"
	"net/mail"
	"net/url"
)

var (
	ErrInvalidURL   = errors.New("invalid URL")
	ErrInvalidEmail = errors.New("invalid email format")
	ErrURLEmptyHost = errors.New("URL has empty host")
	ErrURLBadScheme = errors.New("URL has invalid scheme")
)

// allowedSchemes contains URL schemes that are permitted
var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

// ValidateURL checks if a URL is valid and uses an allowed scheme (http/https)
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return nil // Empty URL is valid (will be skipped)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return errors.Join(ErrInvalidURL, err)
	}

	if u.Host == "" {
		return ErrURLEmptyHost
	}

	if !allowedSchemes[u.Scheme] {
		return errors.Join(ErrURLBadScheme, errors.New(u.Scheme))
	}

	return nil
}

// ValidateEmail validates email format using net/mail ParseAddress
func ValidateEmail(email string) error {
	if email == "" {
		return nil // Empty email is handled by required check
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return errors.Join(ErrInvalidEmail, err)
	}

	// ParseAddress may succeed but return a different address if it parsed
	// only a portion. Ensure the address matches the input.
	if addr.Address != email {
		return ErrInvalidEmail
	}

	return nil
}
