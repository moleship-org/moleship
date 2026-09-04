package middleware

import (
	"net/http"

	"uuid"

	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/domain/persistence"
)

// AdminOnly ensures the authenticated user has admin privileges.
// Must be used after the Auth middleware.
func AdminOnly(userRepo *persistence.UserRepository) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := apiutil.FromRequest(w, r)

			userID, ok := c.Get("user_id").(uuid.UUID)
			if !ok || userID == uuid.Nil() {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := userRepo.FindByID(r.Context(), userID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
