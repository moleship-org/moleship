package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	TokenHash []byte    `json:"token_hash"`
	UserID    uuid.UUID `json:"user_id"`
	IPAddress *string   `json:"ip_address"`
	UserAgent *string   `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
