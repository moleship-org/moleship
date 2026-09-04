package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/api/cookies"
	"github.com/moleship-org/moleship/internal/api/serializer"
	"github.com/moleship-org/moleship/internal/service/auth"
)

type Auth struct {
	authSvc *auth.AuthService
}

func NewAuth(authSvc *auth.AuthService) *Auth {
	return &Auth{authSvc: authSvc}
}

func (h *Auth) Mux(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/login", h.Login)
		})

		r.Group(func(r chi.Router) {
			r.Post("/register", h.Register)
		})

		r.Group(func(r chi.Router) {
			r.Post("/refresh", h.Refresh)
		})

		r.Post("/logout", h.Logout)
	})
}

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	var req serializer.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(http.StatusBadRequest, "invalid login data: "+err.Error())
		return
	}

	token, err := h.authSvc.Login(r.Context(), req.Username, req.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		c.Logger().Info("Invalid credentials for user: %s", req.Username)
		c.Error(http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		c.Logger().Error("error logging in: " + err.Error())
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, cookies.SessionCookie(token))
	c.Set("token", token)
	c.JSON(http.StatusOK, serializer.TokenResponse{Token: token})
}

func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	var req serializer.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(http.StatusBadRequest, "invalid registration data: "+err.Error())
		return
	}

	token, err := h.authSvc.Register(r.Context(), req.Username, req.Email, req.Password)
	if errors.Is(err, auth.ErrUserExists) {
		c.Error(http.StatusConflict, "user already exists")
		return
	}
	if err != nil {
		c.Logger().Error("error registering user: " + err.Error())
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, cookies.SessionCookie(token))
	c.Set("token", token)
	c.JSON(http.StatusCreated, serializer.TokenResponse{Token: token})
}

func (h *Auth) Refresh(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	var req serializer.RefreshRequest
	if err := c.BindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	tokenString := strings.TrimSpace(req.Token)
	if tokenString == "" {
		if cookie, err := r.Cookie(cookies.SessionCookieName); err == nil {
			tokenString = cookie.Value
		}
	}
	if tokenString == "" {
		c.Error(http.StatusBadRequest, "invalid refresh data: token is required")
		return
	}

	token, err := h.authSvc.Refresh(r.Context(), tokenString)
	if errors.Is(err, auth.ErrInvalidToken) {
		c.Error(http.StatusUnauthorized, "invalid or expired token")
		return
	}
	if err != nil {
		c.Logger().Error("error refreshing token: " + err.Error())
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, cookies.SessionCookie(token))
	c.Set("token", token)
	c.JSON(http.StatusOK, serializer.TokenResponse{Token: token})
}

func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	var req serializer.LogoutRequest
	if err := c.BindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	tokenString := strings.TrimSpace(req.Token)
	if tokenString == "" {
		if cookie, err := r.Cookie(cookies.SessionCookieName); err == nil {
			tokenString = cookie.Value
		}
	}
	if tokenString == "" {
		c.Error(http.StatusBadRequest, "invalid logout data: token is required")
		return
	}

	if err := h.authSvc.Logout(r.Context(), tokenString); err != nil {
		c.Logger().Error("error logging out: " + err.Error())
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, cookies.ExpireSessionCookie())
	c.Set("token", nil)
	c.Status(http.StatusNoContent)
}
