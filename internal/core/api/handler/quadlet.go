package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/domain/port"
)

type Quadlet struct {
	quadletSvc port.QuadletService
}

func NewQuadlet(s port.QuadletService) *Quadlet {
	return &Quadlet{
		quadletSvc: s,
	}
}

// --- Quadlet File Operations

func (h *Quadlet) List(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	quadlets, err := h.quadletSvc.List(r.Context())
	if err != nil {
		c.Error(http.StatusInternalServerError, "error trying to get the quadlet list")
		return
	}

	v := &serializer.ListQuadlets{Data: quadlets}
	c.JSON(http.StatusOK, v)
}

func (h *Quadlet) GetByName(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

func (h *Quadlet) Create(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

func (h *Quadlet) Update(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

func (h *Quadlet) ReplaceOrCreate(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

func (h *Quadlet) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

func (h *Quadlet) Mux(r chi.Router) {
	r.Route("/quadlets", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/{name}", h.GetByName)
		r.Post("/", h.Create)
		r.Put("/{name}", h.ReplaceOrCreate)
		r.Patch("/{name}", h.Update)
		r.Delete("/{name}", h.Delete)
	})
}
