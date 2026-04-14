package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/middleware"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/core/service"
	"github.com/moleship-org/moleship/internal/domain/port"
	"golang.org/x/time/rate"
)

type Auth struct {
	authSvc port.AuthService
}

func NewAuth(authSvc port.AuthService) *Auth {
	return &Auth{authSvc: authSvc}
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
	if errors.Is(err, service.ErrInvalidCredentials) {
		c.Error(http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Set("token", token) // Set token in context for potential use in middleware
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
	if errors.Is(err, service.ErrUserExists) {
		c.Error(http.StatusConflict, "user already exists")
		return
	}
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Set("token", token) // Set token in context for potential use in middleware
	c.JSON(http.StatusCreated, serializer.TokenResponse{Token: token})
}

func (h *Auth) Refresh(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	var req serializer.RefreshRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(http.StatusBadRequest, "invalid refresh data: "+err.Error())
		return
	}

	token, err := h.authSvc.Refresh(r.Context(), req.Token)
	if errors.Is(err, service.ErrInvalidToken) {
		c.Error(http.StatusUnauthorized, "invalid or expired token")
		return
	}
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Set("token", token) // Set token in context for potential use in middleware
	c.JSON(http.StatusOK, serializer.TokenResponse{Token: token})
}

func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	var req serializer.LogoutRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(http.StatusBadRequest, "invalid logout data: "+err.Error())
		return
	}

	if err := h.authSvc.Logout(r.Context(), req.Token); err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Set("token", nil) // Clear token from context for potential use in middleware
	c.Status(http.StatusNoContent)
}

func (h *Auth) Mux(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimitByIP(rate.Every(time.Minute), 8))
			r.Post("/login", h.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimitByIP(rate.Every(time.Minute), 4))
			r.Post("/register", h.Register)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimitByIP(rate.Every(time.Minute), 16))
			r.Post("/refresh", h.Refresh)
		})

		r.Post("/logout", h.Logout)
	})
}
