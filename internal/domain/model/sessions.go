package model

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/moleship-org/moleship/internal/adapter/db"
)

type Session struct {
	TokenHash []byte    `json:"token_hash"`
	UserID    uuid.UUID `json:"user_id"`
	IPAddress *string   `json:"ip_address"`
	UserAgent *string   `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Session) Map(row *db.Session) {
	s.TokenHash = row.TokenHash
	s.UserID = uuid.Must(uuid.ParseBytes(row.UserID))
	s.IPAddress = row.IpAddress
	s.UserAgent = row.UserAgent

	t, err := time.Parse(SQLiteTimeLayout, row.ExpiresAt)
	if err != nil {
		log.Fatalf("Error on time.Parse of session ExpiresAt: %s\n", err.Error())
	}
	s.ExpiresAt = t

	t, err = time.Parse(SQLiteTimeLayout, row.CreatedAt)
	if err != nil {
		log.Fatalf("Error on time.Parse of session CreatedAt: %s\n", err.Error())
	}
	s.CreatedAt = t
}
