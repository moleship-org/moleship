package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/moleship-org/moleship/internal/config"
	authsvc "github.com/moleship-org/moleship/internal/services/auth"
	"github.com/moleship-org/moleship/internal/services/auth/authtoken"
	"github.com/moleship-org/moleship/internal/services/auth/cookies"
)

func TestRequireAuthRejectsTokenIssuedBeforePasswordChange(t *testing.T) {
	dir := t.TempDir()
	svc, err := authsvc.NewAuthService(&authsvc.NewAuthServiceParams{
		HostUser:   "tester",
		Dir:        dir,
		BcryptCost: 4,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}

	claims := authtoken.ClaimsFromUser("tester")
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-2 * time.Minute))
	claims.NotBefore = jwt.NewNumericDate(time.Now().Add(-2 * time.Minute))
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(1 * time.Hour))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(config.JWT_SECRET)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if err := svc.SetPassword("new-password"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	h := RequireAuth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/quadlet", nil)
	req.AddCookie(&http.Cookie{Name: cookies.SessionCookieName, Value: tokenStr})
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestRequireAuthAcceptsValidFreshToken(t *testing.T) {
	dir := t.TempDir()
	svc, err := authsvc.NewAuthService(&authsvc.NewAuthServiceParams{
		HostUser:   "tester",
		Dir:        filepath.Clean(dir),
		BcryptCost: 4,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}

	if err := svc.SetPassword("new-password"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	claims := authtoken.ClaimsFromUser("tester")
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(1 * time.Second))
	claims.NotBefore = jwt.NewNumericDate(time.Now())
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(1 * time.Hour))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(config.JWT_SECRET)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	h := RequireAuth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/quadlet", nil)
	req.AddCookie(&http.Cookie{Name: cookies.SessionCookieName, Value: tokenStr})
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, rr.Code)
	}
}
