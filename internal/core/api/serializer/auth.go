package serializer

import (
	"fmt"
	"net/mail"
	"strings"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return fmt.Errorf("username is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		return fmt.Errorf("password is required")
	}
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
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
		return fmt.Errorf("username is required")
	}
	if strings.TrimSpace(r.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		return fmt.Errorf("password is required")
	}
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

type RefreshRequest struct {
	Token string `json:"token"`
}

func (r *RefreshRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}
	return nil
}

type LogoutRequest struct {
	Token string `json:"token"`
}

func (r *LogoutRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}
	return nil
}

type TokenResponse struct {
	Token string `json:"token"`
}
