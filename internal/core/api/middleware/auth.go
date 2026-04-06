package middleware

import (
	"net/http"

	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/service"
	"github.com/moleship-org/moleship/internal/domain/port"
)

func Auth(authSvc port.AuthService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			svc, ok := authSvc.(*service.AuthService)
			if ok {
				if svc.IsOpen() {
					// If open strategy, skip authentication
					next.ServeHTTP(w, r)
					return
				}
			}

			c := apiutil.FromRequest(w, r)

			token := c.RequestHeader().Get("Authorization")
			if token == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Bearer token parsing
			const prefix = "Bearer "
			if len(token) <= len(prefix) || token[:len(prefix)] != prefix {
				http.Error(w, "invalid Authorization header format, it should be Bearer <token>", http.StatusUnauthorized)
				return
			}
			token = token[len(prefix):] // Remove "Bearer " prefix

			userID, err := authSvc.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			c.Set("user_id", userID)
			next.ServeHTTP(w, r)
		})
	}
}
