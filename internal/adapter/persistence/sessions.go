package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

type SessionRepository struct {
	repo Repository
}

func NewSessionRepository(repo Repository) *SessionRepository {
	return &SessionRepository{repo: repo}
}

func (sr *SessionRepository) Save(ctx context.Context, session *model.Session) error {
	err := sr.repo.Querier().CreateSession(ctx, db.CreateSessionParams{
		TokenHash: session.TokenHash,
		UserID:    []byte(session.UserID.String()),
		IpAddress: session.IPAddress,
		UserAgent: session.UserAgent,
		ExpiresAt: session.ExpiresAt.Format(SQLiteTimeLayout),
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return model.ErrSessionExists
		}
	}
	return err
}

func (sr *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash []byte) (*model.Session, error) {
	row, err := sr.repo.Querier().GetSession(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrSessionNotFound
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

func (sr *SessionRepository) FindByUserID(ctx context.Context, userID string) ([]*model.Session, error) {
	rows, err := sr.repo.Querier().GetUserSessions(ctx, []byte(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrSessionNotFound
		}
		return nil, err
	}

	sessions := make([]*model.Session, 0, len(rows))
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

func (sr *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	err := uuid.Validate(userID)
	if err != nil {
		return err
	}

	err = sr.repo.Querier().DeleteAllUserSessions(ctx, []byte(userID))
	return err
}
