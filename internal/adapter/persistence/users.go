package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/moleship-org/moleship/internal/adapter/db"
	"github.com/moleship-org/moleship/internal/domain/model"
)

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
		return nil, err
	}

	user := new(model.User)
	user.Map(&row)

	return user, nil
}

func (ur *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	row, err := ur.repo.Querier().GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	user := new(model.User)
	user.Map(&row)

	return user, nil
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	row, err := ur.repo.Querier().GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	user := new(model.User)
	user.Map(&row)

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
	return err
}

func (ur *UserRepository) List(ctx context.Context, offset int64, limit int64) ([]*model.User, error) {
	rows, err := ur.repo.Querier().ListUsers(ctx, db.ListUsersParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	users := make([]*model.User, len(rows))
	for i, row := range rows {
		user := new(model.User)
		user.Map(&row)
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
