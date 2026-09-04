package persistence

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"uuid"

	"github.com/moleship-org/moleship/database/db"
)

func parseSQLiteTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	layouts := []string{
		time.DateTime,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("unsupported sqlite timestamp: " + value)
}

func parseNullableSQLiteTime(value *string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	parsed, err := parseSQLiteTime(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserInactive    = errors.New("user is inactive")
	ErrUserIsNotAdmin  = errors.New("user is not an admin")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUsersNotFound   = errors.New("users not found")
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	FirstName    *string    `json:"first_name"`
	LastName     *string    `json:"last_name"`
	PasswordHash string     `json:"-"`
	Email        string     `json:"email"`
	IsAdmin      bool       `json:"is_admin"`
	IsActive     bool       `json:"is_active"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"-"`
	UpdatedAt    time.Time  `json:"-"`
	DeletedAt    *time.Time `json:"-"`
}

func MapUser(row *db.User) (u *User, err error) {
	u = new(User)

	u.ID = row.ID
	u.Username = row.Username
	u.FirstName = row.FirstName
	u.LastName = row.LastName
	u.PasswordHash = row.PasswordHash
	u.Email = row.Email
	u.IsAdmin = row.IsAdmin
	u.IsActive = row.IsActive

	u.CreatedAt, err = parseSQLiteTime(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.UpdatedAt, err = parseSQLiteTime(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	u.LastLogin, err = parseNullableSQLiteTime(row.LastLogin)
	if err != nil {
		return nil, err
	}
	u.DeletedAt, err = parseNullableSQLiteTime(row.DeletedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

type UserRepository struct {
	repo Repository
}

func NewUserRepository(repo Repository) *UserRepository {
	return &UserRepository{repo: repo}
}

func (ur *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row, err := ur.repo.Querier().GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user, err := MapUser(&row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	row, err := ur.repo.Querier().GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user, err := MapUser(&row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	row, err := ur.repo.Querier().GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user, err := MapUser(&row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) Save(ctx context.Context, user *User) error {
	err := ur.repo.Querier().CreateUser(ctx, db.CreateUserParams{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Email:        user.Email,
		IsAdmin:      user.IsAdmin,
	})

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrUserExists
		}
	}
	return err
}

func (ur *UserRepository) List(ctx context.Context, offset int64, limit int64) ([]*User, error) {
	rows, err := ur.repo.Querier().ListUsers(ctx, db.ListUsersParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUsersNotFound
		}
		return nil, err
	}

	users := make([]*User, len(rows))
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

func (ur *UserRepository) Update(ctx context.Context, user *User) error {
	err := ur.repo.Querier().UpdateUser(ctx, db.UpdateUserParams{
		ID:           user.ID,
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

func (ur *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return ur.repo.Querier().UpdateUserLastLogin(ctx, id)
}

func (ur *UserRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return ur.repo.Querier().ActivateUser(ctx, id)
}

func (ur *UserRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	return ur.repo.Querier().DeactivateUser(ctx, id)
}

func (ur *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return ur.repo.Querier().SoftDeleteUser(ctx, id)
}

func (ur *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return ur.repo.Querier().HardDeleteUser(ctx, id)
}
