package middleware

import (
	"context"
	"net/http"
	"strings"

	"uuid"

	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/api/authtoken"
	"github.com/moleship-org/moleship/internal/api/cookies"
	"github.com/moleship-org/moleship/internal/domain/config"
	"github.com/moleship-org/moleship/internal/service/auth"
)

func RequireAuth(authSvc *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			cookie, err := r.Cookie(cookies.SessionCookieName)
			if err == nil {
				tokenString = cookie.Value
			} else {
				if config.IsDebug() {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						tokenString = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}
			}

			if tokenString == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userIDStr, err := authSvc.ValidateToken(r.Context(), tokenString)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			claims, err := authtoken.ParseToken(tokenString)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), authtoken.ClaimsKey, claims)
			apiCtx := apiutil.FromRequest(w, r)
			apiCtx.Set("user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
