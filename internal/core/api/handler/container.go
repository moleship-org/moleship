package handler

import (
	"net/http"
	"strings"

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

// List GET /api/containers
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

	res := serializer.ListQuadlet{Data: quadlets}
	if err := ctx.JSON(http.StatusOK, res); err != nil {
		ctx.Error(http.StatusInternalServerError, "Failed to encode response")
	}
}

// GetByName GET /api/containers/{name}
func (h *Container) GetByName(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)

	name := ctx.PathValue("name")
	if strings.TrimSpace(name) == "" {
		ctx.Error(http.StatusBadRequest, "Empty quadlet name")
		return
	}

	quadlet, err := h.containerSvc.GetByName(r.Context(), name)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	res := serializer.GetQuadlet{Data: quadlet}
	if err := ctx.JSON(http.StatusOK, res); err != nil {
		ctx.Error(http.StatusInternalServerError, "Failed to encode response")
	}
}

// PATCH /api/containers/{name}/systemd/start
func (h *Container) Start(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// PATCH /api/containers/{name}/systemd/stop
func (h *Container) Stop(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// PATCH /api/containers/{name}/systemd/restart
func (h *Container) Restart(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// HEAD /api/containers/{name}
func (h *Container) Status(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// GET /api/containers/{name}/logs
func (h *Container) Logs(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

func (h *Container) Mux(m *http.ServeMux) {
	// Get all quadlet entities
	m.HandleFunc("GET /api/containers", h.List)
	// Get one quadlet entity
	m.HandleFunc("GET /api/containers/{name}", h.GetByName)

	// Systemd actions
	m.HandleFunc("PATCH /api/containers/{name}/systemd/start", h.Start)
	m.HandleFunc("PATCH /api/containers/{name}/systemd/stop", h.Stop)
	m.HandleFunc("PATCH /api/containers/{name}/systemd/restart", h.Restart)

	// Status and logs
	m.HandleFunc("HEAD /api/containers/{name}", h.Status)
	m.HandleFunc("GET /api/containers/{name}/logs", h.Logs)
}
