package port

import "context"

type AuthService interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(ctx context.Context, username, email, password string) (string, error)
	Refresh(ctx context.Context, token string) (string, error)
	Logout(ctx context.Context, token string) error
}
