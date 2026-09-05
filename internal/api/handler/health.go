package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Health struct {
	lg *slog.Logger
}

func NewHealth(lg *slog.Logger) *Health {
	r := new(Health)
	r.lg = lg
	return r
}

func (h *Health) Mount(r chi.Router) {
	h.lg.Info("Mounting /health endpoint")

	r.Group(func(r chi.Router) {
		r.Route("/health", func(r chi.Router) {
			r.Get("/", h.Checkhealth)
		})
	})
}

func (h *Health) Checkhealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
