package model

import (
	"time"

	"github.com/google/uuid"
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
