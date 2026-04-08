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

// GetMe godoc
//
//	@Summary		Get current user
//	@Description	Get the profile of the authenticated user
//	@Tags			users
//	@Produce		json
//	@Success		200	{object}	serializer.UserResponse
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		404	{string}	string	"Not found"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/users/me [get]
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

// UpdateMe godoc
//
//	@Summary		Update current user
//	@Description	Update the profile of the authenticated user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		serializer.UpdateUserRequest	true	"Update data"
//	@Success		200		{object}	serializer.UserResponse
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		401		{string}	string	"Unauthorized"
//	@Failure		404		{string}	string	"Not found"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/users/me [put]
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
