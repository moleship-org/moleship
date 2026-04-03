package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/core/service"
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
	c := apiutil.FromRequest(w, r)

	quadlets, err := h.containerSvc.List(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if quadlets == nil {
		c.Status(http.StatusNotFound)
		return
	}

	res := serializer.ListContainer{Data: quadlets}
	if err := c.JSON(http.StatusOK, res); err != nil {
		c.Error(http.StatusInternalServerError, "Failed to encode response")
	}
}

// GetByName GET /api/v1/containers/{name}
func (h *Container) GetByName(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	name := c.PathValue("name")
	if strings.TrimSpace(name) == "" {
		c.Error(http.StatusBadRequest, "Empty container name")
		return
	}

	quadlet, err := h.containerSvc.GetByName(r.Context(), name)
	if errors.Is(err, service.ErrContainertNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
		return
	}

	res := serializer.GetContainer{Data: quadlet}
	if err := c.JSON(http.StatusOK, res); err != nil {
		c.Error(http.StatusInternalServerError, "Failed to encode response")
	}
}

// POST /api/v1/containers/{name}/start
func (h *Container) Start(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	name := c.PathValue("name")
	if strings.TrimSpace(name) == "" {
		c.Error(http.StatusBadRequest, "empty container name")
		return
	}

	err := h.containerSvc.Start(r.Context(), name)
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal error trying to start container")
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /api/v1/containers/{name}/stop
func (h *Container) Stop(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	name := c.PathValue("name")
	if strings.TrimSpace(name) == "" {
		c.Error(http.StatusBadRequest, "empty container name")
		return
	}

	err := h.containerSvc.Stop(r.Context(), name)
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal error trying to stop container")
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /api/v1/containers/{name}/restart
func (h *Container) Restart(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	name := c.PathValue("name")
	if strings.TrimSpace(name) == "" {
		c.Error(http.StatusBadRequest, "empty container name")
		return
	}

	err := h.containerSvc.Restart(r.Context(), name)
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal error trying to restart container")
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/v1/containers/{name}/stats
func (h *Container) Stats(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	name := c.PathValue("name")
	if strings.TrimSpace(name) == "" {
		c.Error(http.StatusBadRequest, "Empty container name")
		return
	}

	stats, err := h.containerSvc.Stats(r.Context(), name)
	if errors.Is(err, service.ErrContainertNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println(stats, err)
		c.Error(http.StatusInternalServerError, "internal error trying to get resources of the container")
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GET /api/v1/containers/{name}/logs
func (h *Container) Logs(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	name := c.PathValue("name")
	if strings.TrimSpace(name) == "" {
		c.Error(http.StatusBadRequest, "empty container name")
		return
	}

	logs, err := h.containerSvc.Logs(r.Context(), name, c.Request().URL.Query())
	if errors.Is(err, service.ErrContainertNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		c.Error(http.StatusInternalServerError, "internal error trying to get logs")
		return
	}
	defer logs.Close()

	c.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Status(http.StatusOK)

	_, copyErr := io.Copy(c.Writer(), logs)
	if copyErr != nil {
		// cannot modify status after header/body write, log internally if required
		c.Logger().Debug("failed to stream logs for %s: %v\n", name, copyErr)
	}
}

func (h *Container) Mux(r chi.Router) {
	r.Route("/containers", func(r chi.Router) {
		// Get containers
		r.Get("/", h.List)
		r.Get("/{name}", h.GetByName)
		// Systemd actions
		r.Post("/{name}/start", h.Start)
		r.Post("/{name}/stop", h.Stop)
		r.Post("/{name}/restart", h.Restart)
		// Status and logs
		r.Get("/{name}/stats", h.Stats)
		r.Get("/{name}/logs", h.Logs)
	})
}
