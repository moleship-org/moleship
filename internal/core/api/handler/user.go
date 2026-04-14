package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/domain/port"
)

type User struct {
	userRepo port.UserRepository
}

func NewUser(userRepo port.UserRepository) *User {
	return &User{userRepo: userRepo}
}

func (h *User) GetMe(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		c.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		c.Error(http.StatusNotFound, "user not found")
		return
	}

	c.JSON(http.StatusOK, serializer.NewUserResponse(user))
}

func (h *User) UpdateMe(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		c.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	var req serializer.UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(http.StatusBadRequest, "invalid update data: "+err.Error())
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		c.Error(http.StatusNotFound, "user not found")
		return
	}

	user.Username = req.Username
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, serializer.NewUserResponse(user))
}

func (h *User) Mux(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/me", h.GetMe)
		r.Put("/me", h.UpdateMe)
	})
}
