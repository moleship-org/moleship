package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
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

func (h *Quadlet) ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

func (h *Quadlet) GetFileByName(w http.ResponseWriter, r *http.Request) {
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
		// Get all quadlet files
		r.Get("/", h.ListFiles)
		// Get one quadlet file
		r.Get("/{name}", h.GetFileByName)
		// Create, Update and Delete
		r.Post("/", h.Create)
		r.Put("/{name}", h.ReplaceOrCreate)
		r.Patch("/{name}", h.Update)
		r.Delete("/{name}", h.Delete)
	})
}
