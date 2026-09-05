package cookies

import (
	"errors"
	"net/http"
	"time"

	"github.com/moleship-org/moleship/internal/config"
)

const (
	SessionCookieName = "moleship_session_token"
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
