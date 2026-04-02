package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/domain/port"
)

type Container struct {
	containerSvc port.ContainerService
}

func NewContainer(s port.ContainerService) *Container {
	return &Container{
		containerSvc: s,
	}
}

// --- Container Entity Operations

// List GET /api/v1/containers
func (h *Container) List(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)

	quadlets, err := h.containerSvc.List(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if quadlets == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	res := serializer.ListContainer{Data: quadlets}
	if err := ctx.JSON(http.StatusOK, res); err != nil {
		ctx.Error(http.StatusInternalServerError, "Failed to encode response")
	}
}

// GetByName GET /api/v1/containers/{name}
func (h *Container) GetByName(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)

	name := chi.URLParam(r, "name")
	if strings.TrimSpace(name) == "" {
		ctx.Error(http.StatusBadRequest, "Empty container name")
		return
	}

	quadlet, err := h.containerSvc.GetByName(r.Context(), name)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	res := serializer.GetContainer{Data: quadlet}
	if err := ctx.JSON(http.StatusOK, res); err != nil {
		ctx.Error(http.StatusInternalServerError, "Failed to encode response")
	}
}

// PATCH /api/v1/containers/{name}/systemd/start
func (h *Container) Start(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// PATCH /api/v1/containers/{name}/systemd/stop
func (h *Container) Stop(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// PATCH /api/v1/containers/{name}/systemd/restart
func (h *Container) Restart(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// HEAD /api/v1/containers/{name}
func (h *Container) Status(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// GET /api/v1/containers/{name}/logs
func (h *Container) Logs(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

func (h *Container) Mux(r chi.Router) {
	r.Route("/containers", func(r chi.Router) {
		// Get containers
		r.Get("/", h.List)
		r.Get("/{name}", h.GetByName)
		// Systemd actions
		r.Patch("/{name}/systemd/start", h.Start)
		r.Patch("/{name}/systemd/stop", h.Stop)
		r.Patch("/{name}/systemd/restart", h.Restart)
		// Status and logs
		r.Head("/{name}", h.Status)
		r.Get("/{name}/logs", h.Logs)
	})
}
