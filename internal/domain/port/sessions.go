package port

import (
	"context"

	"github.com/moleship-org/moleship/internal/domain/model"
)

type SessionRepository interface {
	SessionSaver
	SessionFinder
	SessionDeleter
}

type SessionSaver interface {
	Save(ctx context.Context, session *model.Session) error
}

type SessionFinder interface {
	FindByTokenHash(ctx context.Context, tokenHash []byte) (*model.Session, error)
	FindByUserID(ctx context.Context, userID string) ([]*model.Session, error)
}

type SessionDeleter interface {
	Delete(ctx context.Context, tokenHash []byte) error
	DeleteByUserID(ctx context.Context, userID string) error
}
