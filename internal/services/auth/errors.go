package auth

import "errors"

var (
	ErrNotConfigured = errors.New("auth: no admin password configured")

	ErrInvalidUser = errors.New("auth: invalid user")

	ErrInvalidPassword = errors.New("auth: invalid password")

	ErrCorruptCredentials = errors.New("auth: stored credentials file is corrupt")

	ErrInvalidCredentials = errors.New("auth: invalid credentials")
)
