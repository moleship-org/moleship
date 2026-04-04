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

// List godoc
//
//	@Summary		List quadlets
//	@Description	Get all quadlet files
//	@Tags			quadlets
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	serializer.ListQuadlets
//	@Router			/quadlets [get]
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

// GetByName godoc
//
//	@Summary		Get quadlet by name
//	@Description	Get a quadlet file by name
//	@Tags			quadlets
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string	true	"Quadlet name"
//	@Success		200		{object}	github_com_moleship-org_moleship_internal_domain_model.QuadletFile
//	@Failure		501		{string}	string	"Not implemented"
//	@Router			/quadlets/{name} [get]
func (h *Quadlet) GetByName(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// Create godoc
//
//	@Summary		Create quadlet
//	@Description	Create a new quadlet file
//	@Tags			quadlets
//	@Accept			json
//	@Produce		json
//	@Param			name		path		string	true	"Quadlet name"
//	@Param			quadlet	body			github_com_moleship-org_moleship_internal_domain_model.QuadletFile	true	"Quadlet data"
//	@Success		201			{object}	github_com_moleship-org_moleship_internal_domain_model.QuadletFile
//	@Failure		501			{string}	string	"Not implemented"
//	@Router			/quadlets [post]
func (h *Quadlet) Create(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// Update godoc
//
//	@Summary		Update quadlet
//	@Description	Update an existing quadlet file
//	@Tags			quadlets
//	@Accept			json
//	@Produce		json
//	@Param			name		path		string	true	"Quadlet name"
//	@Param			override	query		bool	false	"Override file"
//	@Param			quadlet	body			github_com_moleship-org_moleship_internal_domain_model.QuadletFile	true	"Quadlet data"
//	@Success		200			{object}	github_com_moleship-org_moleship_internal_domain_model.QuadletFile
//	@Failure		501			{string}	string	"Not implemented"
//	@Router			/quadlets/{name} [patch]
func (h *Quadlet) Update(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// ReplaceOrCreate godoc
//
//	@Summary		Replace or create quadlet
//	@Description	Replace or create a quadlet file
//	@Tags			quadlets
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string				true	"Quadlet name"
//	@Param			quadlet	body		github_com_moleship-org_moleship_internal_domain_model.QuadletFile	true	"Quadlet data"
//	@Success		200		{object}	github_com_moleship-org_moleship_internal_domain_model.QuadletFile
//	@Success		201		{object}	github_com_moleship-org_moleship_internal_domain_model.QuadletFile
//	@Failure		501		{string}	string				"Not implemented"
//	@Router			/quadlets/{name} [put]
func (h *Quadlet) ReplaceOrCreate(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	ctx.Status(http.StatusNotImplemented)
}

// Delete godoc
//
//	@Summary		Delete quadlet
//	@Description	Delete a quadlet file
//	@Tags			quadlets
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string	true	"Quadlet name"
//	@Success		204		{string}	string	"No content"
//	@Failure		501		{string}	string	"Not implemented"
//	@Router			/quadlets/{name} [delete]
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
