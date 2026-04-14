package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/domain/port"
)

type Admin struct {
	userRepo port.UserRepository
}

func NewAdmin(userRepo port.UserRepository) *Admin {
	return &Admin{userRepo: userRepo}
}

func (h *Admin) ListUsers(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	offset := parseIntQuery(c.QueryParam("offset"), 0)
	limit := parseIntQuery(c.QueryParam("limit"), 20)

	users, err := h.userRepo.List(r.Context(), offset, limit)
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	total, err := h.userRepo.Count(r.Context())
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	data := make([]serializer.UserResponse, len(users))
	for i, u := range users {
		data[i] = serializer.NewUserResponse(u)
	}

	c.JSON(http.StatusOK, serializer.ListUsersResponse{
		Data:   data,
		Offset: offset,
		Limit:  limit,
		Total:  total,
	})
}

func (h *Admin) GetUser(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	id := c.PathValue("id")
	if id == "" {
		c.Error(http.StatusBadRequest, "missing user ID")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), id)
	if err != nil {
		c.Error(http.StatusNotFound, "user not found")
		return
	}

	c.JSON(http.StatusOK, serializer.NewUserResponse(user))
}

func (h *Admin) UpdateUser(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	id := c.PathValue("id")
	if id == "" {
		c.Error(http.StatusBadRequest, "missing user ID")
		return
	}

	var req serializer.AdminUpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(http.StatusBadRequest, "invalid update data: "+err.Error())
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), id)
	if err != nil {
		c.Error(http.StatusNotFound, "user not found")
		return
	}

	user.Username = req.Username
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	user.IsAdmin = req.IsAdmin
	user.IsActive = req.IsActive

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, serializer.NewUserResponse(user))
}

func (h *Admin) ActivateUser(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	id := c.PathValue("id")
	if id == "" {
		c.Error(http.StatusBadRequest, "missing user ID")
		return
	}

	if err := h.userRepo.Activate(r.Context(), id); err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Admin) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	id := c.PathValue("id")
	if id == "" {
		c.Error(http.StatusBadRequest, "missing user ID")
		return
	}

	if err := h.userRepo.Deactivate(r.Context(), id); err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Admin) SoftDeleteUser(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	id := c.PathValue("id")
	if id == "" {
		c.Error(http.StatusBadRequest, "missing user ID")
		return
	}

	if err := h.userRepo.SoftDelete(r.Context(), id); err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Admin) HardDeleteUser(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	id := c.PathValue("id")
	if id == "" {
		c.Error(http.StatusBadRequest, "missing user ID")
		return
	}

	if err := h.userRepo.HardDelete(r.Context(), id); err != nil {
		c.Error(http.StatusInternalServerError, "internal server error")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Admin) Mux(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Get("/", h.ListUsers)
			r.Get("/{id}", h.GetUser)
			r.Put("/{id}", h.UpdateUser)
			r.Post("/{id}/activate", h.ActivateUser)
			r.Post("/{id}/deactivate", h.DeactivateUser)
			r.Delete("/{id}", h.SoftDeleteUser)
			r.Delete("/{id}/hard", h.HardDeleteUser)
		})
	})
}

func parseIntQuery(val string, def int64) int64 {
	if val == "" {
		return def
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil || n < 0 {
		return def
	}
	return n
}
