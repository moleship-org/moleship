package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/core/api/serializer"
	"github.com/moleship-org/moleship/internal/domain/model"
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
//	@Failure		404		{string}	string	"Not found"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/quadlets/{name} [get]
func (h *Quadlet) GetByName(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)
	name := chi.URLParam(r, "name")

	qf, err := h.quadletSvc.Get(r.Context(), name)
	if err != nil {
		c.Error(http.StatusNotFound, "quadlet not found")
		return
	}

	c.JSON(http.StatusOK, qf)
}

// Create godoc
//
//	@Summary		Create quadlet
//	@Description	Create a new quadlet file
//	@Tags			quadlets
//	@Accept			json
//	@Produce		json
//	@Param			quadlet	body			github_com_moleship-org_moleship_internal_domain_model.QuadletFile	true	"Quadlet data"
//	@Success		201			{object}	github_com_moleship-org_moleship_internal_domain_model.QuadletFile
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/quadlets [post]
func (h *Quadlet) Create(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)

	var qf model.QuadletFile
	if err := json.NewDecoder(r.Body).Decode(&qf); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	if qf.Name == "" {
		c.Error(http.StatusBadRequest, "quadlet name is required")
		return
	}

	if err := h.quadletSvc.Create(r.Context(), qf.Name, &qf); err != nil {
		c.Error(http.StatusInternalServerError, "error creating quadlet")
		return
	}

	c.JSON(http.StatusCreated, qf)
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
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		404			{string}	string	"Not found"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/quadlets/{name} [patch]
func (h *Quadlet) Update(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)
	name := c.PathValue("name")
	override, _ := strconv.ParseBool(c.QueryParam("override"))

	var qf model.QuadletFile
	if err := json.NewDecoder(r.Body).Decode(&qf); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.quadletSvc.Update(r.Context(), override, name, &qf); err != nil {
		c.Error(http.StatusInternalServerError, "error updating quadlet")
		return
	}

	c.JSON(http.StatusOK, qf)
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
//	@Failure		400		{string}	string				"Bad request"
//	@Failure		500		{string}	string				"Internal server error"
//	@Router			/quadlets/{name} [put]
func (h *Quadlet) ReplaceOrCreate(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)
	name := chi.URLParam(r, "name")

	var qf model.QuadletFile
	if err := json.NewDecoder(r.Body).Decode(&qf); err != nil {
		c.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	exists, err := h.quadletSvc.Exists(r.Context(), name)
	if err != nil {
		c.Error(http.StatusInternalServerError, "error checking if quadlet exists")
		return
	}

	if exists {
		if err := h.quadletSvc.Update(r.Context(), true, name, &qf); err != nil {
			c.Error(http.StatusInternalServerError, "error updating quadlet")
			return
		}
		c.JSON(http.StatusOK, qf)
	} else {
		qf.Name = name
		if err := h.quadletSvc.Create(r.Context(), name, &qf); err != nil {
			c.Error(http.StatusInternalServerError, "error creating quadlet")
			return
		}
		c.JSON(http.StatusCreated, qf)
	}
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
//	@Failure		404		{string}	string	"Not found"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/quadlets/{name} [delete]
func (h *Quadlet) Delete(w http.ResponseWriter, r *http.Request) {
	c := apiutil.FromRequest(w, r)
	name := c.PathValue("name")

	if err := h.quadletSvc.Delete(r.Context(), name); err != nil {
		c.Error(http.StatusNotFound, "quadlet not found")
		return
	}

	c.Status(http.StatusNoContent)
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
