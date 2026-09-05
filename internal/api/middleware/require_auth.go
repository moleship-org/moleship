package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/moleship-org/moleship/internal/config"
	authsvc "github.com/moleship-org/moleship/internal/services/auth"
	"github.com/moleship-org/moleship/internal/services/auth/authtoken"
	"github.com/moleship-org/moleship/internal/services/auth/cookies"
)

func RequireAuth(authService *authsvc.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookies.SessionCookieName)
			if err != nil || cookie.Value == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims := &authtoken.Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, http.ErrAbortHandler
				}
				return config.JWT_SECRET, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			if authService != nil {
				changedAt, err := authService.ChangedAt()
				if err != nil {
					if errors.Is(err, authsvc.ErrNotConfigured) {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}

				issuedAt, err := claims.GetIssuedAt()
				if err != nil || issuedAt == nil || issuedAt.Time.Before(changedAt) {
					http.Error(w, "invalid or expired token", http.StatusUnauthorized)
					return
				}
			}

			ctx := context.WithValue(r.Context(), authtoken.ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
