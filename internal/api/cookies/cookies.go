package cookies

import (
	"net/http"
	"time"

	"github.com/moleship-org/moleship/internal/domain/config"
)

const (
	SessionCookieName = "moleship_session_token"
	SessionDuration   = 7 * 24 * time.Hour
)

func SessionCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.IsProduction(),
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
		Secure:   config.IsProduction(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
}
