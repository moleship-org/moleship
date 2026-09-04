package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"uuid"

	"github.com/moleship-org/moleship/database/db"
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

func MapSession(row *db.Session) (s *Session, err error) {
	s = new(Session)

	s.UserID = row.UserID
	s.TokenHash = row.TokenHash
	s.IPAddress = row.IpAddress
	s.UserAgent = row.UserAgent

	s.ExpiresAt, err = parseSQLiteTime(row.ExpiresAt)
	if err != nil {
		return nil, err
	}
	s.CreatedAt, err = parseSQLiteTime(row.CreatedAt)
	if err != nil {
		return nil, err
	}

	return s, nil
}

type SessionRepository struct {
	repo Repository
}

func NewSessionRepository(repo Repository) *SessionRepository {
	return &SessionRepository{repo: repo}
}

func (sr *SessionRepository) Save(ctx context.Context, session *Session) error {
	err := sr.repo.Querier().CreateSession(ctx, db.CreateSessionParams{
		TokenHash: session.TokenHash,
		UserID:    session.UserID,
		IpAddress: session.IPAddress,
		UserAgent: session.UserAgent,
		ExpiresAt: session.ExpiresAt,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrSessionExists
		}
	}
	return err
}

func (sr *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash []byte) (*Session, error) {
	row, err := sr.repo.Querier().GetSession(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	session, err := MapSession(&db.Session{
		TokenHash: row.TokenHash,
		UserID:    row.UserID,
		IpAddress: row.IpAddress,
		UserAgent: row.UserAgent,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (sr *SessionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error) {
	rows, err := sr.repo.Querier().GetUserSessions(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	sessions := make([]*Session, 0, len(rows))
	for _, row := range rows {
		session, err := MapSession(&db.Session{
			TokenHash: row.TokenHash,
			UserID:    row.UserID,
			IpAddress: row.IpAddress,
			UserAgent: row.UserAgent,
			ExpiresAt: row.ExpiresAt,
			CreatedAt: row.CreatedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("error mapping session for user %s: %w", userID, err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (sr *SessionRepository) Delete(ctx context.Context, tokenHash []byte) error {
	err := sr.repo.Querier().DeleteSession(ctx, tokenHash)
	return err
}

func (sr *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return sr.repo.Querier().DeleteAllUserSessions(ctx, userID)
}
