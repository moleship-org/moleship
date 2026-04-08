package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

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
