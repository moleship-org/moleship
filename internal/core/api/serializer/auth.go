package serializer

import (
	"fmt"
	"net/mail"
	"strings"
)

var (
	ErrEmptyPassword = fmt.Errorf("password must not be empty")
	ErrShortPassword = fmt.Errorf("password must be at least 8 characters long")
	ErrEmptyToken    = fmt.Errorf("token is required")
	ErrInvalidToken  = fmt.Errorf("invalid token format")
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return ErrEmptyUsername
	}
	if strings.TrimSpace(r.Password) == "" {
		return ErrEmptyPassword
	}
	if len(r.Password) < 8 {
		return ErrShortPassword
	}
	return nil
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return ErrEmptyUsername
	}
	if strings.TrimSpace(r.Email) == "" {
		return ErrEmptyEmail
	}
	if strings.TrimSpace(r.Password) == "" {
		return ErrEmptyPassword
	}
	if len(r.Password) < 8 {
		return ErrShortPassword
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

type RefreshRequest struct {
	Token string `json:"token"`
}

func (r *RefreshRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return ErrEmptyToken
	}
	return nil
}

type LogoutRequest struct {
	Token string `json:"token"`
}

func (r *LogoutRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return ErrEmptyToken
	}
	return nil
}

type TokenResponse struct {
	Token string `json:"token"`
}
