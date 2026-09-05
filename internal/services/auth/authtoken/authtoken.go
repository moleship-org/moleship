package authtoken

import (
	"context"
	"time"

	"uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/moleship-org/moleship/internal/services/auth/cookies"
)

type contextKey string

const ClaimsKey contextKey = "claims"

const DefaultIssuer = "moleship-api"

type Claims struct {
	jwt.RegisteredClaims
	User string `json:"sub"`
}

func ClaimsFromUser(user string) *Claims {
	return &Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cookies.SessionDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    DefaultIssuer,
			ID:        uuid.NewV7().String(),
		},
	}
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*Claims)
	return claims, ok
}

func UserFromContext(ctx context.Context) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return "", false
	}
	return claims.User, true
}
