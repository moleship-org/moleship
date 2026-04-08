package serializer

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/moleship-org/moleship/internal/domain/model"
)

var (
	ErrEmptyUsername = fmt.Errorf("username is required")
	ErrEmptyEmail    = fmt.Errorf("email is required")
	ErrInvalidEmail  = fmt.Errorf("invalid email format")
)

type UserResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     string  `json:"email"`
	IsAdmin   bool    `json:"is_admin"`
	IsActive  bool    `json:"is_active"`
}

func NewUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		IsAdmin:   u.IsAdmin,
		IsActive:  u.IsActive,
	}
}

type ListUsersResponse struct {
	Data   []UserResponse `json:"data"`
	Offset int64          `json:"offset"`
	Limit  int64          `json:"limit"`
	Total  int64          `json:"total"`
}

type UpdateUserRequest struct {
	Username  string  `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     string  `json:"email"`
}

func (r *UpdateUserRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return ErrEmptyUsername
	}
	if strings.TrimSpace(r.Email) == "" {
		return ErrEmptyEmail
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

type AdminUpdateUserRequest struct {
	Username  string  `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     string  `json:"email"`
	IsAdmin   bool    `json:"is_admin"`
	IsActive  bool    `json:"is_active"`
}

func (r *AdminUpdateUserRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return ErrEmptyUsername
	}
	if strings.TrimSpace(r.Email) == "" {
		return ErrEmptyEmail
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}
