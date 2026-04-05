package persistence

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moleship-org/moleship/internal/adapter/db"
	"github.com/moleship-org/moleship/internal/domain/model"
)

func MapSession(row *db.Session) (s *model.Session, err error) {
	s = new(model.Session)

	id, err := uuid.ParseBytes(row.UserID)
	if err != nil {
		return s, fmt.Errorf("error on uuid.ParseBytes of session UserID: %w", err)
	}
	s.UserID = id

	s.TokenHash = row.TokenHash
	s.IPAddress = row.IpAddress
	s.UserAgent = row.UserAgent

	t, err := time.Parse(SQLiteTimeLayout, row.ExpiresAt)
	if err != nil {
		return s, fmt.Errorf("error on time.Parse of session ExpiresAt: %w", err)
	}
	s.ExpiresAt = t

	t, err = time.Parse(SQLiteTimeLayout, row.CreatedAt)
	if err != nil {
		return s, fmt.Errorf("error on time.Parse of session CreatedAt: %w", err)
	}
	s.CreatedAt = t

	return s, nil
}
