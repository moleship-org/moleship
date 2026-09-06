package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/moleship-org/moleship/internal/services/auth/cookies"
)

const csrfHeaderName = "X-CSRF-Token"

func RequireCSRF() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !requiresCSRFCheck(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			sessionCookie, err := r.Cookie(cookies.SessionCookieName)
			if err != nil || sessionCookie.Value == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			csrfCookie, err := r.Cookie(cookies.CSRFCookieName)
			if err != nil || csrfCookie.Value == "" {
				http.Error(w, "missing csrf token", http.StatusForbidden)
				return
			}

			headerToken := r.Header.Get(csrfHeaderName)
			if headerToken == "" {
				http.Error(w, "missing csrf token", http.StatusForbidden)
				return
			}

			if subtle.ConstantTimeCompare([]byte(csrfCookie.Value), []byte(headerToken)) != 1 {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func requiresCSRFCheck(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
