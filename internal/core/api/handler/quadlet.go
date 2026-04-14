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
	c := apiutil.FromRequest(w, r)
	name := chi.URLParam(r, "name")

	qf, err := h.quadletSvc.Get(r.Context(), name)
	if err != nil {
		c.Error(http.StatusNotFound, "quadlet not found")
		return
	}

	c.JSON(http.StatusOK, qf)
}

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
