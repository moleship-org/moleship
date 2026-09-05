package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moleship-org/moleship/internal/services/auth/cookies"
)

func TestRequireCSRFSkipsSafeMethods(t *testing.T) {
	h := RequireCSRF()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/systemd/units/demo/status", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestRequireCSRFRejectsUnsafeMethodWithoutToken(t *testing.T) {
	h := RequireCSRF()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/systemd/daemon-reload", nil)
	req.AddCookie(&http.Cookie{Name: cookies.SessionCookieName, Value: "session"})
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestRequireCSRFAcceptsMatchingToken(t *testing.T) {
	h := RequireCSRF()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/quadlet/containers/demo", nil)
	req.AddCookie(&http.Cookie{Name: cookies.SessionCookieName, Value: "session"})
	req.AddCookie(&http.Cookie{Name: cookies.CSRFCookieName, Value: "csrf-token"})
	req.Header.Set("X-CSRF-Token", "csrf-token")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, rr.Code)
	}
}
