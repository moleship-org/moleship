package port

import (
	"context"
)

type AuthService interface {
	// Login authenticates a user with username and password and returns a token if successful
	Login(ctx context.Context, username, password string) (string, error)
	// Register creates a new user and returns a token for the created user
	Register(ctx context.Context, username, email, password string) (string, error)
	// Refresh generates a new token for the provided token if it's valid
	Refresh(ctx context.Context, token string) (string, error)
	// Logout invalidates the provided token
	Logout(ctx context.Context, token string) error
	// ValidateToken validates the provided token and returns the associated user ID if valid
	ValidateToken(ctx context.Context, token string) (string, error)
}
