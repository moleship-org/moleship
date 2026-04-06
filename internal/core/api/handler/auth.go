package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/core/service"
	"github.com/moleship-org/moleship/internal/domain/port"
)

type Auth struct {
	authSvc port.AuthService
}

func NewAuth(authSvc port.AuthService) *Auth {
	return &Auth{authSvc: authSvc}
}

// Login godoc
//
//	@Summary		Login
//	@Description	Authenticate with username and password, returns a session token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		serializer.LoginRequest	true	"Login credentials"
//	@Success		200		{object}	serializer.TokenResponse	"Session token"
//	@Failure		400		{string}	string			"Bad request"
//	@Failure		401		{string}	string			"Invalid credentials"
//	@Failure		500		{string}	string			"Internal server error"
//	@Router			/auth/login [post]
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

// Register godoc
//
//	@Summary		Register
//	@Description	Create a new user account, returns a session token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		serializer.RegisterRequest	true	"Registration data"
//	@Success		201		{object}	serializer.TokenResponse	"Session token"
//	@Failure		400		{string}	string			"Bad request"
//	@Failure		409		{string}	string			"User already exists"
//	@Failure		500		{string}	string			"Internal server error"
//	@Router			/auth/register [post]
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

// Refresh godoc
//
//	@Summary		Refresh token
//	@Description	Exchange a valid session token for a new one
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		serializer.RefreshRequest	true	"Current session token"
//	@Success		200		{object}	serializer.TokenResponse	"New session token"
//	@Failure		400		{string}	string			"Bad request"
//	@Failure		401		{string}	string			"Invalid or expired token"
//	@Failure		500		{string}	string			"Internal server error"
//	@Router			/auth/refresh [post]
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

// Logout godoc
//
//	@Summary		Logout
//	@Description	Invalidate a session token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		serializer.LogoutRequest	true	"Session token to invalidate"
//	@Success		204		{string}	string			"No content"
//	@Failure		400		{string}	string			"Bad request"
//	@Failure		500		{string}	string			"Internal server error"
//	@Router			/auth/logout [post]
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
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)
	})
}
