package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionExists   = errors.New("session already exists")
	ErrInvalidToken    = errors.New("invalid token")
)

type Session struct {
	TokenHash []byte    `json:"-"`
	UserID    uuid.UUID `json:"user_id"`
	IPAddress *string   `json:"ip_address"`
	UserAgent *string   `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"-"`
}
