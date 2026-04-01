package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/domain/port"
)

type Quadlet struct {
	service port.QuadletService
}

func NewQuadlet(s port.QuadletService) *Quadlet {
	return &Quadlet{
		service: s,
	}
}

// List GET /api/quadlets
func (h *Quadlet) List(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)

	quadlets, err := h.service.List(r.Context())
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

// GetByName GET /api/quadlets/{name}
func (h *Quadlet) GetByName(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)

	name := ctx.PathValue("name")
	fmt.Println("Name", name)
	if strings.TrimSpace(name) == "" {
		ctx.Error(http.StatusBadRequest, "Empty quadlet name")
		return
	}

	quadlet, err := h.service.GetByName(r.Context(), name)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	res := serializer.GetQuadlet{Data: quadlet}
	if err := ctx.JSON(http.StatusOK, res); err != nil {
		ctx.Error(http.StatusInternalServerError, "Failed to encode response")
	}
}

func (h *Quadlet) Mux(m *http.ServeMux) {
	m.HandleFunc("GET /api/quadlets", h.List)
	m.HandleFunc("GET /api/quadlets/{name}", h.GetByName)
	// m.HandleFunc("/api/quadlets/start/", ht.Start)
	// m.HandleFunc("/api/quadlets/stop/", ht.Stop)
	// m.HandleFunc("/api/quadlets/restart/", ht.Restart)
}
