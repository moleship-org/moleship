package handler

import (
	"net/http"

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

func (h *Quadlet) Mux(m *http.ServeMux) {
	// Get all quadlet files
	m.HandleFunc("GET /api/quadlets", h.ListFiles)
	// Get one quadlet file
	m.HandleFunc("GET /api/quadlets/{name}", h.GetFileByName)

	// Create, Update and Delete
	m.HandleFunc("POST /api/quadlets", h.Create)
	m.HandleFunc("PUT /api/quadlets/{name}", h.ReplaceOrCreate)
	m.HandleFunc("PATCH /api/quadlets/{name}", h.Update)
	m.HandleFunc("DELETE /api/quadlets/{name}", h.Delete)
}
