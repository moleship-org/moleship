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

func MapUser(row *db.User) (u *model.User, err error) {
	u = new(model.User)

	id, err := uuid.ParseBytes(row.ID)
	if err != nil {
		return u, fmt.Errorf("error on uuid.ParseBytes of user ID: %w", err)
	}
	u.ID = id

	u.Username = row.Username
	u.FirstName = row.FirstName
	u.LastName = row.LastName
	u.PasswordHash = row.PasswordHash
	u.Email = row.Email
	u.IsAdmin = row.IsAdmin
	u.IsActive = row.IsActive

	if row.LastLogin != nil {
		t, err := time.Parse(SQLiteTimeLayout, *row.LastLogin)
		if err != nil {
			return u, fmt.Errorf("error on time.Parse of user LastLogin: %w", err)
		}
		u.LastLogin = &t
	}

	t, err := time.Parse(SQLiteTimeLayout, row.CreatedAt)
	if err != nil {
		return u, fmt.Errorf("error on time.Parse of user CreatedAt: %w", err)
	}
	u.CreatedAt = t

	t, err = time.Parse(SQLiteTimeLayout, row.UpdatedAt)
	if err != nil {
		return u, fmt.Errorf("error on time.Parse of user UpdatedAt: %w", err)
	}
	u.UpdatedAt = t

	if row.DeletedAt != nil {
		t, err := time.Parse(SQLiteTimeLayout, *row.DeletedAt)
		if err != nil {
			return u, fmt.Errorf("error on time.Parse of user DeletedAt: %w", err)
		}
		u.DeletedAt = &t
	}

	return u, nil
}

type UserRepository struct {
	repo Repository
}

func NewUserRepository(repo Repository) *UserRepository {
	return &UserRepository{repo: repo}
}

func (ur *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	err := uuid.Validate(id)
	if err != nil {
		return nil, err
	}

	row, err := ur.repo.Querier().GetUser(ctx, []byte(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	user, err := MapUser(&row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	row, err := ur.repo.Querier().GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	user, err := MapUser(&row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	row, err := ur.repo.Querier().GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	user, err := MapUser(&row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) Save(ctx context.Context, user *model.User) error {
	err := ur.repo.Querier().CreateUser(ctx, db.CreateUserParams{
		ID:           []byte(user.ID.String()),
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Email:        user.Email,
		IsAdmin:      user.IsAdmin,
	})

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return model.ErrUserExists
		}
	}
	return err
}

func (ur *UserRepository) List(ctx context.Context, offset int64, limit int64) ([]*model.User, error) {
	rows, err := ur.repo.Querier().ListUsers(ctx, db.ListUsersParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUsersNotFound
		}
		return nil, err
	}

	users := make([]*model.User, len(rows))
	for i, row := range rows {
		user, err := MapUser(&row)
		if err != nil {
			return nil, err
		}
		users[i] = user
	}

	return users, nil
}

func (ur *UserRepository) Count(ctx context.Context) (int64, error) {
	count, err := ur.repo.Querier().CountUsers(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (ur *UserRepository) Update(ctx context.Context, user *model.User) error {
	err := ur.repo.Querier().UpdateUser(ctx, db.UpdateUserParams{
		ID:           []byte(user.ID.String()),
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		PasswordHash: user.PasswordHash,
		Email:        user.Email,
		IsAdmin:      user.IsAdmin,
		IsActive:     user.IsActive,
	})
	return err
}

func (ur *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	err := uuid.Validate(id)
	if err != nil {
		return err
	}

	err = ur.repo.Querier().UpdateUserLastLogin(ctx, []byte(id))
	return err
}

func (ur *UserRepository) Activate(ctx context.Context, id string) error {
	err := uuid.Validate(id)
	if err != nil {
		return err
	}

	err = ur.repo.Querier().ActivateUser(ctx, []byte(id))
	return err
}

func (ur *UserRepository) Deactivate(ctx context.Context, id string) error {
	err := uuid.Validate(id)
	if err != nil {
		return err
	}

	err = ur.repo.Querier().DeactivateUser(ctx, []byte(id))
	return err
}

func (ur *UserRepository) SoftDelete(ctx context.Context, id string) error {
	err := uuid.Validate(id)
	if err != nil {
		return err
	}

	err = ur.repo.Querier().SoftDeleteUser(ctx, []byte(id))
	return err
}

func (ur *UserRepository) HardDelete(ctx context.Context, id string) error {
	err := uuid.Validate(id)
	if err != nil {
		return err
	}

	err = ur.repo.Querier().HardDeleteUser(ctx, []byte(id))
	return err
}
