package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalidUsername or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionExists      = errors.New("session already exists")
)
