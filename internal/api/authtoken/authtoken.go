package authtoken

import (
	"context"
	"errors"
	"strings"
	"time"

	"uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/moleship-org/moleship/internal/api/cookies"
	"github.com/moleship-org/moleship/internal/domain/config"
)

type contextKey string

const ClaimsKey contextKey = "claims"

const DefaultIssuer = "moleship-backend-api"

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"sub"`
}

func ClaimsFromUserID(userID uuid.UUID) *Claims {
	return &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cookies.SessionDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    DefaultIssuer,
			ID:        uuid.New().String(),
		},
	}
}

func NewSignedToken(userID uuid.UUID) (string, error) {
	claims := ClaimsFromUserID(userID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.Current().JWTSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("token is required")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return config.Current().JWTSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*Claims)
	return claims, ok
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return uuid.Nil(), false
	}
	return claims.UserID, true
}
