package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/domain/port"
)

type Auth struct {
	userRepo port.UserRepository
}

func NewAuth(userRepo port.UserRepository) *Auth {
	return &Auth{userRepo: userRepo}
}

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)
	c.Status(http.StatusNotImplemented)
}

func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)
	c.Status(http.StatusNotImplemented)
}

func (h *Auth) Refresh(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)
	c.Status(http.StatusNotImplemented)
}

func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)
	c.Status(http.StatusNotImplemented)
}

func (h *Auth) Mux(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)
	})
}
