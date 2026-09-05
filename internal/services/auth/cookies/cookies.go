package cookies

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/moleship-org/moleship/internal/config"
)

const (
	SessionCookieName = "moleship_session_token"
	CSRFCookieName    = "moleship_csrf_token"
	SessionDuration   = 24 * time.Hour
)

var ErrNoSession = errors.New("cookies: no session cookie present")

func SessionCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.IsReleaseMode(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(SessionDuration.Seconds()),
	}
}

func CSRFCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   config.IsReleaseMode(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(SessionDuration.Seconds()),
	}
}

func ExpireSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   config.IsReleaseMode(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
}

func ExpireCSRFCookie() *http.Cookie {
	return &http.Cookie{
		Name:     CSRFCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   config.IsReleaseMode(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	}
}

func ReadSession(r *http.Request) (string, error) {
	c, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", ErrNoSession
	}
	if c.Value == "" {
		return "", ErrNoSession
	}
	return c.Value, nil
}

func NewCSRFToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("cookies: failed to generate csrf token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
