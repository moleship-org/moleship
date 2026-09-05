package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/config"
	"github.com/moleship-org/moleship/internal/services/auth"
	"github.com/moleship-org/moleship/internal/services/auth/authtoken"
	"github.com/moleship-org/moleship/internal/services/auth/cookies"
)

type Auth struct {
	lg  *slog.Logger
	svc *auth.AuthService
}

func NewAuth(l *slog.Logger, s *auth.AuthService) *Auth {
	return &Auth{lg: l, svc: s}
}

func (h *Auth) Mount(r chi.Router) {
	h.lg.Info("Mounting /auth endpoint")

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/logout", h.Logout)
		r.Get("/status", h.Status)
		r.Get("/session", h.Session)
	})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// POST /auth/login
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	c := apiutil.From(w, r)

	var body loginRequest
	if err := c.BindJSON(&body); err != nil {
		h.lg.Error("auth bind json error", slog.String("error", err.Error()))
		c.Error(http.StatusBadRequest, "invalid payload")
		return
	}

	if !h.svc.IsConfigured() {
		c.Error(http.StatusPreconditionFailed, "instance not configured yet")
		return
	}

	if err := h.svc.Verify(body.Username, body.Password); err != nil {
		if !errors.Is(err, auth.ErrInvalidCredentials) && !errors.Is(err, auth.ErrNotConfigured) {
			h.lg.Error("auth verify error", slog.String("error", err.Error()))
		}
		c.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	claims := authtoken.ClaimsFromUser(body.Username)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSig, err := token.SignedString(config.JWT_SECRET)
	if err != nil {
		h.lg.Error("auth sign token error", slog.String("error", err.Error()))
		c.Error(http.StatusInternalServerError, "internal error")
		return
	}

	c.Cookie(http.StatusOK, cookies.SessionCookie(tokenSig))
}

// POST /auth/logout
func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	c := apiutil.From(w, r)
	c.Cookie(http.StatusNoContent, cookies.ExpireSessionCookie())
}

// GET /auth/status
func (h *Auth) Status(w http.ResponseWriter, r *http.Request) {
	c := apiutil.From(w, r)

	c.JSON(http.StatusOK, map[string]bool{
		"configured": h.svc.IsConfigured(),
	})
}

// GET /auth/session
func (h *Auth) Session(w http.ResponseWriter, r *http.Request) {
	c := apiutil.From(w, r)

	tokenStr, err := cookies.ReadSession(r)
	if err != nil {
		c.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	claims := &authtoken.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return config.JWT_SECRET, nil
	})
	if err != nil || !token.Valid {
		c.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	changedAt, err := h.svc.ChangedAt()
	if err != nil {
		// No hay contraseña configurada (o se acaba de resetear vía
		// password_init sin que el proceso haya vuelto a levantar
		// creds en memoria todavía): cualquier token existente es
		// inválido por definición.
		c.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	issuedAt, err := claims.GetIssuedAt()
	if err != nil || issuedAt == nil || issuedAt.Time.Before(changedAt) {
		c.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"username": claims.User,
	})
}
