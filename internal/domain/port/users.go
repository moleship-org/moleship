package port

import (
	"context"

	"github.com/moleship-org/moleship/internal/domain/model"
)

type UserRepository interface {
	UserFinder
	UserSaver
	UserUpdater
	UserDeleter
	UserLister
}

type UserSaver interface {
	Save(ctx context.Context, user *model.User) error
}

type UserFinder interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type UserLister interface {
	List(ctx context.Context, offset int64, limit int64) ([]*model.User, error)
	Count(ctx context.Context) (int64, error)
}

type UserUpdater interface {
	Update(ctx context.Context, user *model.User) error
	UpdateLastLogin(ctx context.Context, id string) error
}

type UserDeleter interface {
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}
